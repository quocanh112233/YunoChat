package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles WebSocket upgrade requests
type Handler struct {
	hub            *Hub
	jwtSecret      string
	upgrader       websocket.Upgrader
	allowedOrigins []string
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, cfg *config.Config) *Handler {
	return &Handler{
		hub:       hub,
		jwtSecret: cfg.JWT.AccessSecret,
		allowedOrigins: cfg.Server.AllowedOrigins,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // Allow requests with no Origin header (e.g. some apps/tools)
				}

				// If "*" is in allowed origins, allow all
				for _, allowed := range cfg.Server.AllowedOrigins {
					if allowed == "*" {
						return true
					}
					if allowed == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// HandleUpgrade handles WebSocket upgrade requests
// Endpoint: GET /v1/ws?token=<JWT>
func (h *Handler) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter (WebSocket doesn't support custom headers)
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing token parameter", http.StatusUnauthorized)
		return
	}

	// Parse and validate JWT
	userID, err := h.parseToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already writes error response
		return
	}

	// Create new client
	client := NewClient(h.hub, conn, userID)

	// Register client with hub
	h.hub.register <- client

	// Send welcome event
	h.sendWelcome(client, userID)

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// parseToken validates the JWT token and returns the user ID
func (h *Handler) parseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, jwt.ErrInvalidKey
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, jwt.ErrInvalidKey
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// sendWelcome sends the connected event to the client
func (h *Handler) sendWelcome(client *Client, userID uuid.UUID) {
	payload := ConnectedPayload{
		UserID:     userID.String(),
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}

	// Send immediately before writePump starts
	msg := BaseMessage{
		Event:   EventConnected,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// Write directly since writePump hasn't started yet
	client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	client.conn.WriteMessage(websocket.TextMessage, data)
}
