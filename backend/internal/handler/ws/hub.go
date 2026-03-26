package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// PostgreSQL NOTIFY payload limit is 8000 bytes, use 7500 as safety margin
	maxPayloadSize = 7500
	// Grace period before marking user as OFFLINE
	gracePeriodDuration = 60 * time.Second
	// Channel buffer size for client send channel
	sendChannelBuffer = 256
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients: userID -> list of connections (multi-tab support)
	clients map[uuid.UUID][]*Client
	mu      sync.RWMutex

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Grace period timers for presence
	gracePeriods map[uuid.UUID]*time.Timer
	gpMu         sync.Mutex

	// Dedicated PostgreSQL connection for LISTEN/NOTIFY
	listenConn *pgx.Conn
	dbURL      string

	// DB pool for friendship checks and other queries
	pool *pgxpool.Pool

	// Context for shutting down
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub creates a new Hub instance
func NewHub(dbListenURL string, pool *pgxpool.Pool) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:      make(map[uuid.UUID][]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		gracePeriods: make(map[uuid.UUID]*time.Timer),
		dbURL:        dbListenURL,
		pool:         pool,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// IsFriends checks if two users are friends (friendship.status = ACCEPTED)
func (h *Hub) IsFriends(ctx context.Context, userA, userB uuid.UUID) bool {
	if h.pool == nil {
		return true // fallback: allow if no DB (should not happen in prod)
	}
	var status string
	err := h.pool.QueryRow(ctx,
		`SELECT status FROM friendships
		 WHERE (requester_id = $1 AND addressee_id = $2)
		    OR (requester_id = $2 AND addressee_id = $1)
		 LIMIT 1`,
		userA, userB,
	).Scan(&status)
	if err != nil {
		return false
	}
	return status == "ACCEPTED"
}

// GetConversationParticipants returns user IDs of conversation participants (excluding caller)
func (h *Hub) GetConversationParticipants(ctx context.Context, conversationID string, excludeUserID uuid.UUID) []uuid.UUID {
	if h.pool == nil {
		return nil
	}
	rows, err := h.pool.Query(ctx,
		`SELECT user_id FROM conversation_participants
		 WHERE conversation_id = $1
		   AND left_at IS NULL
		   AND user_id != $2`,
		conversationID, excludeUserID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []uuid.UUID
	for rows.Next() {
		var id pgtype.UUID
		if err := rows.Scan(&id); err == nil {
			result = append(result, id.Bytes)
		}
	}
	return result
}

// IsMember checks if a user is a member of a conversation
func (h *Hub) IsMember(ctx context.Context, conversationID string, userID uuid.UUID) bool {
	if h.pool == nil {
		return true
	}
	var exists bool
	err := h.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM conversation_participants
			WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL
		)`,
		conversationID, userID,
	).Scan(&exists)
	return err == nil && exists
}

// Run starts the Hub's goroutines
func (h *Hub) Run() {
	go h.handleClients()
	go h.ListenLoop()
}

// handleClients processes register/unregister requests
func (h *Hub) handleClients() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
			// Cancel grace period if exists (user reconnected)
			h.cancelGracePeriod(client.userID)

		case client := <-h.unregister:
			h.removeClient(client)
			// Start grace period if no other connections
			h.startGracePeriod(client.userID)
		}
	}
}

// addClient adds a client to the hub
func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update DB status to ONLINE
	if h.pool != nil {
		_, err := h.pool.Exec(context.Background(),
			"UPDATE users SET status = 'ONLINE' WHERE id = $1",
			client.userID,
		)
		if err != nil {
			log.Printf("Error updating user status to ONLINE: %v", err)
		}
	}

	h.clients[client.userID] = append(h.clients[client.userID], client)
	count := len(h.clients[client.userID])
	h.mu.Unlock()

	log.Printf("Client registered: userID=%s, total connections=%d", client.userID, count)

	// Broadcast presence update (only if first connection)
	if count == 1 {
		h.broadcastPresence(client.userID, "ONLINE")
	}
}

// removeClient removes a client from the hub
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	connections := h.clients[client.userID]
	for i, c := range connections {
		if c == client {
			// Remove this connection from the slice
			h.clients[client.userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// If no more connections for this user, clean up the map entry
	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)
	}

	close(client.send)
	log.Printf("Client unregistered: userID=%s, remaining connections=%d", client.userID, len(h.clients[client.userID]))
}

// hasActiveConnections checks if user has any active connections
func (h *Hub) hasActiveConnections(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// startGracePeriod starts a timer to mark user as offline after 60s
func (h *Hub) startGracePeriod(userID uuid.UUID) {
	// Don't start grace period if user has other active connections
	if h.hasActiveConnections(userID) {
		return
	}

	h.gpMu.Lock()
	defer h.gpMu.Unlock()

	// Cancel existing timer if any
	if timer, exists := h.gracePeriods[userID]; exists {
		timer.Stop()
	}

	// Start new grace period timer
	timer := time.AfterFunc(gracePeriodDuration, func() {
		// Update DB status to OFFLINE
		if h.pool != nil {
			_, err := h.pool.Exec(context.Background(),
				"UPDATE users SET status = 'OFFLINE', last_seen_at = NOW() WHERE id = $1",
				userID,
			)
			if err != nil {
				log.Printf("Error updating user status to OFFLINE: %v", err)
			}
		}

		h.broadcastPresence(userID, "OFFLINE")
		h.gpMu.Lock()
		delete(h.gracePeriods, userID)
		h.gpMu.Unlock()
	})

	h.gracePeriods[userID] = timer
	log.Printf("Grace period started for userID=%s", userID)
}

// cancelGracePeriod cancels the grace period timer when user reconnects
func (h *Hub) cancelGracePeriod(userID uuid.UUID) {
	h.gpMu.Lock()
	defer h.gpMu.Unlock()

	if timer, exists := h.gracePeriods[userID]; exists {
		timer.Stop()
		delete(h.gracePeriods, userID)
		log.Printf("Grace period cancelled for userID=%s (user reconnected)", userID)
	}
}

// broadcastPresence broadcasts presence update to relevant users
func (h *Hub) broadcastPresence(userID uuid.UUID, status string) {
	payload := PresenceUpdatePayload{
		UserID: userID.String(),
		Status: status,
	}
	if status == "OFFLINE" {
		now := time.Now().UTC().Format(time.RFC3339)
		payload.LastSeenAt = now
	}

	msg := BaseMessage{
		Event:   EventPresenceUpdate,
		Payload: payload,
	}

	// TODO: Broadcast only to users who share conversations with this user
	// For now, broadcast to all connected clients (MVP simplification)
	h.broadcast(msg)
}

// broadcast sends a message to all connected clients
func (h *Hub) broadcast(msg BaseMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for _, client := range clients {
			select {
			case client.send <- data:
			default:
				// Client's send channel is full, close it
				close(client.send)
			}
		}
	}
}

// ListenLoop listens for PostgreSQL NOTIFY events on 'chat_events' channel
func (h *Hub) ListenLoop() {
	ctx := h.ctx

	// Create dedicated connection for LISTEN
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := h.connectAndListen(ctx); err != nil {
			log.Printf("LISTEN connection error: %v, reconnecting in 5s...", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}
}

// connectAndListen establishes connection and listens for notifications
func (h *Hub) connectAndListen(ctx context.Context) error {
	// Create dedicated connection (not from pool)
	conn, err := pgx.Connect(ctx, h.dbURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	h.listenConn = conn

	// Listen on chat_events channel
	_, err = conn.Exec(ctx, "LISTEN chat_events")
	if err != nil {
		return err
	}

	log.Println("LISTEN connection established on 'chat_events' channel")

	for {
		// Wait for notification (blocking call)
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled, clean exit
			}
			return err
		}

		// Dispatch in goroutine to avoid blocking the listen loop
		go h.dispatch(notification.Payload)
	}
}

// dispatch parses notification and sends to appropriate clients
func (h *Hub) dispatch(rawPayload string) {
	var event ChatEvent
	if err := json.Unmarshal([]byte(rawPayload), &event); err != nil {
		log.Printf("Error parsing notification payload: %v", err)
		return
	}

	// If payload is large, we only got message_id, need to query DB
	if event.MessageID != nil && event.Data == nil {
		// TODO: Query DB to get full message data
		// For now, skip - this requires repository integration
		log.Printf("Large payload case: message_id=%s (needs DB query)", *event.MessageID)
		return
	}

	// Build WebSocket message
	msg := BaseMessage{
		Event:   event.Type,
		Payload: event.Data,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	// Send to recipients
	if len(event.RecipientIDs) > 0 {
		// Send to specific recipients
		for _, recipientID := range event.RecipientIDs {
			userID, err := uuid.Parse(recipientID)
			if err != nil {
				continue
			}
			h.sendToUser(userID, data)
		}
	} else {
		// Broadcast to all participants of the conversation
		// For now, broadcast to all connected clients (MVP simplification)
		h.broadcast(msg)
	}
}

// sendToUser sends data to all connections of a specific user
func (h *Hub) sendToUser(userID uuid.UUID, data []byte) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			// Channel full, skip this client
		}
	}
}

// Close gracefully shuts down the Hub
// It closes all client send channels, which triggers their writePumps to send close frames.
func (h *Hub) Close() error {
	log.Println("Closing WebSocket Hub...")

	// Cancel context to stop ListenLoop
	h.cancel()

	// Close listen connection
	if h.listenConn != nil {
		h.listenConn.Close(context.Background())
	}

	// Close all client send channels
	h.mu.Lock()
	for userID, clients := range h.clients {
		for _, client := range clients {
			// We only close the channel. The writePump will see this,
			// send the CloseMessage and then close the connection.
			close(client.send)
		}
		delete(h.clients, userID)
	}
	h.mu.Unlock()

	// Stop all grace period timers
	h.gpMu.Lock()
	for userID, timer := range h.gracePeriods {
		timer.Stop()
		delete(h.gracePeriods, userID)
	}
	h.gpMu.Unlock()

	log.Println("WebSocket Hub closed")
	return nil
}

// GetClientCount returns the number of connected clients (for debugging)
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

func (h *Hub) handleMarkRead(userID string, conversationID string) {
	// Update DB last_read_at
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userUUID, _ := uuid.Parse(userID)
	convUUID, _ := uuid.Parse(conversationID)

	userPgID := pgtype.UUID{Bytes: userUUID, Valid: true}
	convPgID := pgtype.UUID{Bytes: convUUID, Valid: true}

	_, err := h.pool.Exec(ctx,
		"UPDATE conversation_participants SET last_read_at = NOW() WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL",
		convPgID, userPgID,
	)
	if err != nil {
		log.Printf("Error updating last_read_at: %v", err)
		return
	}

	// Get latest message ID for read receipt
	var lastMsgID pgtype.UUID
	err = h.pool.QueryRow(ctx,
		"SELECT id FROM messages WHERE conversation_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1",
		convPgID,
	).Scan(&lastMsgID)

	if err == nil && lastMsgID.Valid {
		// Broadcast read receipt to other participants
		participants, _ := h.getParticipants(ctx, convPgID)
		recipientIDs := make([]string, 0, len(participants))
		for _, p := range participants {
			pID, _ := uuid.FromBytes(p.Bytes[:])
			if pID.String() != userID {
				recipientIDs = append(recipientIDs, pID.String())
			}
		}

		if len(recipientIDs) > 0 {
			msgID, _ := uuid.FromBytes(lastMsgID.Bytes[:])
			event := map[string]interface{}{
				"type": "read_receipt",
				"data": map[string]interface{}{
					"conversation_id":      conversationID,
					"reader_id":            userID,
					"last_read_message_id": msgID.String(),
					"read_at":              time.Now().Format(time.RFC3339),
				},
				"recipient_ids": recipientIDs,
			}
			h.broadcastEvent(event)
		}
	}
}

func (h *Hub) handleTyping(userID string, conversationID string, isTyping bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	convUUID, _ := uuid.Parse(conversationID)
	convPgID := pgtype.UUID{Bytes: convUUID, Valid: true}

	// Get participants to broadcast
	participants, _ := h.getParticipants(ctx, convPgID)
	recipientIDs := make([]string, 0, len(participants))
	for _, p := range participants {
		pID, _ := uuid.FromBytes(p.Bytes[:])
		if pID.String() != userID {
			recipientIDs = append(recipientIDs, pID.String())
		}
	}

	if len(recipientIDs) > 0 {
		// Get user info for typing indicator
		var displayName string
		_ = h.pool.QueryRow(ctx, "SELECT display_name FROM users WHERE id = $1", pgtype.UUID{Bytes: uuid.MustParse(userID), Valid: true}).Scan(&displayName)

		event := map[string]interface{}{
			"type": "user_typing",
			"data": map[string]interface{}{
				"conversation_id": conversationID,
				"user": map[string]interface{}{
					"id":           userID,
					"display_name": displayName,
				},
				"is_typing": isTyping,
			},
			"recipient_ids": recipientIDs,
		}
		h.broadcastEvent(event)
	}
}

func (h *Hub) getParticipants(ctx context.Context, convID pgtype.UUID) ([]pgtype.UUID, error) {
	rows, err := h.pool.Query(ctx, "SELECT user_id FROM conversation_participants WHERE conversation_id = $1 AND left_at IS NULL", convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []pgtype.UUID
	for rows.Next() {
		var id pgtype.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (h *Hub) broadcastEvent(event map[string]interface{}) {
	jsonPayload, _ := json.Marshal(event)
	_, err := h.pool.Exec(context.Background(), "SELECT pg_notify('chat_events', $1)", string(jsonPayload))
	if err != nil {
		log.Printf("Error broadcasting event: %v", err)
	}
}
