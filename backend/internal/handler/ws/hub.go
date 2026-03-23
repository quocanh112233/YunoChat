package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	// Context for shutting down
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub creates a new Hub instance
func NewHub(dbListenURL string) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:      make(map[uuid.UUID][]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		gracePeriods: make(map[uuid.UUID]*time.Timer),
		dbURL:        dbListenURL,
		ctx:          ctx,
		cancel:       cancel,
	}
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

	h.clients[client.userID] = append(h.clients[client.userID], client)
	log.Printf("Client registered: userID=%s, total connections=%d", client.userID, len(h.clients[client.userID]))
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
	if event.RecipientIDs != nil && len(event.RecipientIDs) > 0 {
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
func (h *Hub) Close() error {
	log.Println("Closing WebSocket Hub...")

	// Cancel context to stop ListenLoop
	h.cancel()

	// Close listen connection
	if h.listenConn != nil {
		h.listenConn.Close(context.Background())
	}

	// Close all client connections
	h.mu.Lock()
	for userID, clients := range h.clients {
		for _, client := range clients {
			close(client.send)
			client.conn.Close()
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
