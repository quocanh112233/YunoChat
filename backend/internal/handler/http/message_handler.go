package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"backend/internal/pkg/response"
	msguc "backend/internal/usecase/message"
)

// MessageHandler handles message-related HTTP endpoints
type MessageHandler struct {
	sendUC       *msguc.SendMessageUseCase
	listUC       *msguc.ListMessagesUseCase
	softDeleteUC *msguc.SoftDeleteUseCase
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	sendUC *msguc.SendMessageUseCase,
	listUC *msguc.ListMessagesUseCase,
	softDeleteUC *msguc.SoftDeleteUseCase,
) *MessageHandler {
	return &MessageHandler{
		sendUC:       sendUC,
		listUC:       listUC,
		softDeleteUC: softDeleteUC,
	}
}

// RegisterRoutes registers message routes
func (h *MessageHandler) RegisterRoutes(r chi.Router) {
	r.Get("/conversations/{id}/messages", h.ListMessages)
	r.Post("/conversations/{id}/messages", h.SendMessage)
	r.Delete("/messages/{id}", h.DeleteMessage)
}

// ListMessages handles GET /v1/conversations/:id/messages
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		response.Err(w, http.StatusBadRequest, "MISSING_ID", "Conversation ID is required")
		return
	}

	// Parse query params
	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 50 {
				limit = 50
			}
		}
	}

	var beforeID, beforeTime *string
	if bid := r.URL.Query().Get("before_id"); bid != "" {
		beforeID = &bid
	}
	if bt := r.URL.Query().Get("before_time"); bt != "" {
		beforeTime = &bt
	}

	req := msguc.ListMessagesRequest{
		UserID:         userID,
		ConversationID: conversationID,
		Limit:          int32(limit),
		BeforeID:       beforeID,
		BeforeTime:     beforeTime,
	}

	resp, err := h.listUC.Execute(r.Context(), req)
	if err != nil {
		switch err {
		case msguc.ErrNotMember:
			response.Err(w, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this conversation")
		case msguc.ErrInvalidMessageCursor:
			response.Err(w, http.StatusBadRequest, "INVALID_CURSOR", "Invalid cursor parameters")
		default:
			response.Err(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list messages")
		}
		return
	}

	// Build meta
	meta := &response.Meta{
		HasMore: resp.Meta.HasMore,
	}
	if resp.Meta.NextCursor != nil {
		meta.Cursor = resp.Meta.NextCursor.BeforeID
	}

	response.OKWithMeta(w, http.StatusOK, resp.Messages, meta)
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Body         string              `json:"body,omitempty"`
	Type         string              `json:"type"` // TEXT or ATTACHMENT
	Attachment   *msguc.AttachmentData `json:"attachment,omitempty"`
	ClientTempID string              `json:"client_temp_id,omitempty"`
}

// SendMessage handles POST /v1/conversations/:id/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		response.Err(w, http.StatusBadRequest, "MISSING_ID", "Conversation ID is required")
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// Validate message type
	if req.Type != "TEXT" && req.Type != "ATTACHMENT" {
		response.Err(w, http.StatusBadRequest, "INVALID_TYPE", "Message type must be TEXT or ATTACHMENT")
		return
	}

	ucReq := msguc.SendMessageRequest{
		SenderID:       userID,
		ConversationID: conversationID,
		Body:           req.Body,
		Type:           req.Type,
		Attachment:     req.Attachment,
		ClientTempID:   req.ClientTempID,
	}

	resp, err := h.sendUC.Execute(r.Context(), ucReq)
	if err != nil {
		switch err {
		case msguc.ErrInvalidMessage:
			response.Err(w, http.StatusBadRequest, "INVALID_MESSAGE", "Invalid message content")
		case msguc.ErrAttachmentRequired:
			response.Err(w, http.StatusBadRequest, "ATTACHMENT_REQUIRED", "Attachment data required for ATTACHMENT type")
		case msguc.ErrNotFriends:
			response.Err(w, http.StatusForbidden, "NOT_FRIENDS", "You are not friends with this user - cannot send DM")
		case msguc.ErrNotMember:
			response.Err(w, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this conversation")
		default:
			response.Err(w, http.StatusInternalServerError, "SEND_FAILED", "Failed to send message")
		}
		return
	}

	response.OK(w, http.StatusCreated, resp)
}

// DeleteMessage handles DELETE /v1/messages/:id
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	messageID := chi.URLParam(r, "id")
	if messageID == "" {
		response.Err(w, http.StatusBadRequest, "MISSING_ID", "Message ID is required")
		return
	}

	req := msguc.SoftDeleteRequest{
		UserID:    userID,
		MessageID: messageID,
	}

	resp, err := h.softDeleteUC.Execute(r.Context(), req)
	if err != nil {
		switch err {
		case msguc.ErrNotSender:
			response.Err(w, http.StatusForbidden, "NOT_SENDER", "Only the sender can delete this message")
		case msguc.ErrMessageNotFound:
			response.Err(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Message not found")
		case msguc.ErrAlreadyDeleted:
			response.Err(w, http.StatusBadRequest, "ALREADY_DELETED", "Message has already been deleted")
		default:
			response.Err(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete message")
		}
		return
	}

	response.OK(w, http.StatusOK, map[string]interface{}{
		"message_id": resp.MessageID,
		"deleted":    resp.Deleted,
	})
}
