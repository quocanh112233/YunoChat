package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the Hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// User ID associated with this connection
	userID uuid.UUID

	// Conversations this client has joined
	rooms map[string]bool
	mu    sync.RWMutex
}

// NewClient creates a new Client instance
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, sendChannelBuffer),
		userID: userID,
		rooms:  make(map[string]bool),
	}
}

// readPump pumps messages from the websocket connection to the Hub
// The application runs readPump in a per-connection goroutine
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.handleMessage(message)
	}
}

// writePump pumps messages from the Hub to the websocket connection
// A goroutine running writePump is started for each connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel, send close frame with message
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(data []byte) {
	var msg BaseMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("INVALID_MESSAGE", "Invalid message format", "")
		return
	}

	switch msg.Event {
	case EventPing:
		c.handlePing()
	case EventJoinConversation:
		c.handleJoinConversation(msg.Payload)
	case EventLeaveConversation:
		c.handleLeaveConversation(msg.Payload)
	case EventSendMessage:
		c.handleSendMessage(msg.Payload, msg.ID)
	case EventTyping:
		c.handleTyping(msg.Payload)
	case EventMarkRead:
		c.handleMarkRead(msg.Payload)
	default:
		c.sendError("UNKNOWN_EVENT", "Unknown event type: "+msg.Event, msg.Event)
	}
}

// handlePing responds with pong
func (c *Client) handlePing() {
	payload := PongPayload{
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
	c.sendEvent(EventPong, payload)
}

// handleJoinConversation adds client to a conversation room
func (c *Client) handleJoinConversation(payload interface{}) {
	data, _ := json.Marshal(payload)
	var join JoinConversationPayload
	if err := json.Unmarshal(data, &join); err != nil {
		c.sendError("VALIDATION_ERROR", "Invalid payload", EventJoinConversation)
		return
	}

	c.mu.Lock()
	c.rooms[join.ConversationID] = true
	c.mu.Unlock()

	log.Printf("User %s joined conversation %s", c.userID, join.ConversationID)
}

// handleLeaveConversation removes client from a conversation room
func (c *Client) handleLeaveConversation(payload interface{}) {
	data, _ := json.Marshal(payload)
	var leave LeaveConversationPayload
	if err := json.Unmarshal(data, &leave); err != nil {
		c.sendError("VALIDATION_ERROR", "Invalid payload", EventLeaveConversation)
		return
	}

	c.mu.Lock()
	delete(c.rooms, leave.ConversationID)
	c.mu.Unlock()

	log.Printf("User %s left conversation %s", c.userID, leave.ConversationID)
}

// handleSendMessage processes send_message event
// Validates membership (all types) and friendship (DM only) before accepting message
func (c *Client) handleSendMessage(payload interface{}, _ string) {
	data, _ := json.Marshal(payload)
	var sendMsg SendMessagePayload
	if err := json.Unmarshal(data, &sendMsg); err != nil {
		c.sendError("VALIDATION_ERROR", "Invalid payload", EventSendMessage)
		return
	}

	if sendMsg.ConversationID == "" {
		c.sendError("VALIDATION_ERROR", "conversation_id is required", EventSendMessage)
		return
	}

	// Use a short-lived context for DB checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Verify caller is a member of the conversation
	if !c.hub.IsMember(ctx, sendMsg.ConversationID, c.userID) {
		c.sendError("FORBIDDEN", "You are not a member of this conversation", EventSendMessage)
		return
	}

	// 2. For DM conversations, also verify active friendship
	// Query conversation type and the other participant
	if c.hub.pool != nil {
		var convType string
		var otherUserID uuid.UUID
		row := c.hub.pool.QueryRow(ctx,
			`SELECT c.type, cp.user_id
			 FROM conversations c
			 JOIN conversation_participants cp ON cp.conversation_id = c.id
			 WHERE c.id = $1
			   AND cp.user_id != $2
			   AND cp.left_at IS NULL
			 LIMIT 1`,
			sendMsg.ConversationID, c.userID,
		)
		if err := row.Scan(&convType, &otherUserID); err == nil && convType == "DM" {
			// DM: check friendship status
			if !c.hub.IsFriends(ctx, c.userID, otherUserID) {
				c.sendError("FORBIDDEN", "Bạn không thể gửi tin nhắn cho người này", EventSendMessage)
				return
			}
		}
	}

	// Send acknowledgment back to sender
	ack := MessageSentPayload{
		ClientTempID: sendMsg.ClientTempID,
		MessageID:    "", // Full message persistence is handled by HTTP endpoint
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		Status:       "SENT",
	}
	c.sendEvent(EventMessageSent, ack)

	log.Printf("User %s sent message to conversation %s", c.userID, sendMsg.ConversationID)
}

// handleTyping broadcasts typing indicator
func (c *Client) handleTyping(payload interface{}) {
	data, _ := json.Marshal(payload)
	var typing TypingPayload
	if err := json.Unmarshal(data, &typing); err != nil {
		return
	}

	c.hub.handleTyping(c.userID.String(), typing.ConversationID, typing.IsTyping)
}

// handleMarkRead marks messages as read
func (c *Client) handleMarkRead(payload interface{}) {
	data, _ := json.Marshal(payload)
	var markRead MarkReadPayload
	if err := json.Unmarshal(data, &markRead); err != nil {
		c.sendError("VALIDATION_ERROR", "Invalid payload", EventMarkRead)
		return
	}

	c.hub.handleMarkRead(c.userID.String(), markRead.ConversationID)
}

// sendEvent sends an event to the client
func (c *Client) sendEvent(event string, payload interface{}) {
	msg := BaseMessage{
		Event:   event,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		// Channel full, drop message
	}
}

// sendError sends an error event to the client
func (c *Client) sendError(code, message, refEvent string) {
	payload := ErrorPayload{
		Code:     code,
		Message:  message,
		RefEvent: refEvent,
	}
	c.sendEvent(EventError, payload)
}

// IsInRoom checks if client has joined a specific conversation
func (c *Client) IsInRoom(conversationID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rooms[conversationID]
}
