package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/pkg/response"
	convuc "backend/internal/usecase/conversation"

	"github.com/go-chi/chi/v5"
)

// ConversationHandler handles conversation-related HTTP endpoints
type ConversationHandler struct {
	listUC        *convuc.ListConversationsUseCase
	getUC         *convuc.GetConversationUseCase
	createGroupUC *convuc.CreateGroupUseCase
	markReadUC    *convuc.MarkReadUseCase
	kickMemberUC  *convuc.KickMemberUseCase
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(
	listUC *convuc.ListConversationsUseCase,
	getUC *convuc.GetConversationUseCase,
	createGroupUC *convuc.CreateGroupUseCase,
	markReadUC *convuc.MarkReadUseCase,
	kickMemberUC *convuc.KickMemberUseCase,
) *ConversationHandler {
	return &ConversationHandler{
		listUC:        listUC,
		getUC:         getUC,
		createGroupUC: createGroupUC,
		markReadUC:    markReadUC,
		kickMemberUC:  kickMemberUC,
	}
}

// RegisterRoutes registers conversation routes
func (h *ConversationHandler) RegisterRoutes(r chi.Router) {
	r.Get("/conversations", h.ListConversations)
	r.Get("/conversations/{id}", h.GetConversation)
	r.Post("/conversations/groups", h.CreateGroup)
	r.Patch("/conversations/{id}/read", h.MarkAsRead)
	r.Delete("/conversations/{id}/members/{user_id}", h.KickMember)
}

// GetConversation handles GET /v1/conversations/{id}
func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.getUC.Execute(r.Context(), convuc.GetConversationRequest{
		UserID:         userID,
		ConversationID: conversationID,
	})
	if err != nil {
		if err == convuc.ErrConversationNotFound {
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Conversation not found")
			return
		}
		response.Err(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get conversation details")
		return
	}

	response.OK(w, http.StatusOK, resp)
}

// ListConversations handles GET /v1/conversations
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
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

	req := convuc.ListConversationsRequest{
		UserID:     userID,
		Limit:      int32(limit),
		BeforeID:   beforeID,
		BeforeTime: beforeTime,
	}

	resp, err := h.listUC.Execute(r.Context(), req)
	if err != nil {
		if err == convuc.ErrInvalidCursor {
			response.Err(w, http.StatusBadRequest, "INVALID_CURSOR", "Invalid cursor parameters")
			return
		}
		response.Err(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list conversations")
		return
	}

	// Build meta
	meta := &response.Meta{
		HasMore: resp.Meta.HasMore,
	}
	if resp.Meta.NextCursor != nil {
		meta.Cursor = resp.Meta.NextCursor.BeforeID
	}

	response.OKWithMeta(w, http.StatusOK, resp.Conversations, meta)
}

// CreateGroupRequest represents the request body for creating a group
type CreateGroupRequest struct {
	Name           string   `json:"name"`
	ParticipantIDs []string `json:"participant_ids"`
}

// CreateGroup handles POST /v1/conversations/groups
func (h *ConversationHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// Validate minimum participants (creator + 2 others = 3 total)
	if len(req.ParticipantIDs) < 2 {
		response.Err(w, http.StatusBadRequest, "MIN_PARTICIPANTS", "Group requires at least 2 other participants (3 total)")
		return
	}

	ucReq := convuc.CreateGroupRequest{
		CreatorID:      userID,
		Name:           req.Name,
		ParticipantIDs: req.ParticipantIDs,
	}

	resp, err := h.createGroupUC.Execute(r.Context(), ucReq)
	if err != nil {
		switch err {
		case convuc.ErrInvalidName:
			response.Err(w, http.StatusBadRequest, "INVALID_NAME", "Group name is required")
		case convuc.ErrMinParticipants:
			response.Err(w, http.StatusBadRequest, "MIN_PARTICIPANTS", "Group requires at least 3 participants")
		default:
			response.Err(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create group")
		}
		return
	}

	response.OK(w, http.StatusCreated, resp)
}

// MarkReadRequest represents the request body for marking as read
type MarkReadRequest struct {
	LastMessageID string `json:"last_message_id,omitempty"`
}

// MarkAsRead handles PATCH /v1/conversations/{id}/read
func (h *ConversationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	var req MarkReadRequest
	// Try to decode body, but it's optional
	_ = json.NewDecoder(r.Body).Decode(&req)

	ucReq := convuc.MarkReadRequest{
		UserID:         userID,
		ConversationID: conversationID,
		LastMessageID:  req.LastMessageID,
	}

	resp, err := h.markReadUC.Execute(r.Context(), ucReq)
	if err != nil {
		switch err {
		case convuc.ErrNotMember:
			response.Err(w, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this conversation")
		case convuc.ErrMessageNotFound:
			response.Err(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Message not found")
		default:
			response.Err(w, http.StatusInternalServerError, "MARK_READ_FAILED", "Failed to mark as read")
		}
		return
	}

	response.OK(w, http.StatusOK, map[string]interface{}{
		"last_read_message_id": resp.LastReadMessageID,
		"read_at":              resp.ReadAt,
	})
}

// KickMember handles DELETE /v1/conversations/{id}/members/{user_id}
// Only group admins can kick other members.
func (h *ConversationHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value("user_id").(string)
	if !ok || callerID == "" {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	conversationID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "user_id")
	if conversationID == "" || targetUserID == "" {
		response.Err(w, http.StatusBadRequest, "MISSING_PARAMS", "Missing conversation or user ID")
		return
	}

	err := h.kickMemberUC.Execute(r.Context(), convuc.KickMemberRequest{
		CallerID:       callerID,
		ConversationID: conversationID,
		TargetUserID:   targetUserID,
	})
	if err != nil {
		switch err {
		case convuc.ErrNotAdmin:
			response.Err(w, http.StatusForbidden, "NOT_ADMIN", "Only admin can kick members")
		case convuc.ErrNotMember:
			response.Err(w, http.StatusNotFound, "NOT_MEMBER", "User is not a member")
		case convuc.ErrCannotSelf:
			response.Err(w, http.StatusBadRequest, "CANNOT_KICK_SELF", "Cannot kick yourself")
		default:
			response.Err(w, http.StatusInternalServerError, "KICK_FAILED", "Failed to kick member")
		}
		return
	}

	response.OK(w, http.StatusOK, map[string]interface{}{"success": true})
}
