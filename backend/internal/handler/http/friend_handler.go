package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/pkg/response"
	"backend/internal/usecase/friendship"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FriendHandler handles friend-related HTTP requests
type FriendHandler struct {
	sendRequestUC    *friendship.SendRequestUseCase
	respondRequestUC *friendship.RespondRequestUseCase
	unfriendUC       *friendship.UnfriendUseCase
	listUC           *friendship.ListUseCase
}

// RegisterRoutes registers friend routes
func (h *FriendHandler) RegisterRoutes(r chi.Router) {
	r.Route("/friends", func(r chi.Router) {
		r.Get("/", h.ListFriends)
		r.Route("/requests", func(r chi.Router) {
			r.Post("/", h.SendRequest)
			r.Get("/received", h.ListPendingReceived)
			r.Get("/sent", h.ListPendingSent)
			r.Patch("/{id}", h.RespondRequest)
			r.Delete("/{id}", h.CancelRequest)
		})
		r.Delete("/{id}", h.Unfriend)
	})

	r.Get("/users/search", h.SearchUsers)
}

// NewFriendHandler creates a new FriendHandler
func NewFriendHandler(
	sendRequestUC *friendship.SendRequestUseCase,
	respondRequestUC *friendship.RespondRequestUseCase,
	unfriendUC *friendship.UnfriendUseCase,
	listUC *friendship.ListUseCase,
) *FriendHandler {
	return &FriendHandler{
		sendRequestUC:    sendRequestUC,
		respondRequestUC: respondRequestUC,
		unfriendUC:       unfriendUC,
		listUC:           listUC,
	}
}

// SendRequest handles POST /v1/friends/requests
func (h *FriendHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}

	requesterID := r.Context().Value("user_id").(uuid.UUID)

	resp, err := h.sendRequestUC.Execute(r.Context(), friendship.SendRequestRequest{
		RequesterID: requesterID,
		AddresseeID: toUserID,
	})
	if err != nil {
		switch err {
		case friendship.ErrSelfRequest:
			response.Err(w, http.StatusBadRequest, "VALIDATION_ERROR", "Không thể gửi lời mời cho chính mình")
		case friendship.ErrUserNotFound:
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Người dùng không tồn tại")
		case friendship.ErrDuplicateRequest:
			response.Err(w, http.StatusConflict, "CONFLICT", "Lời mời đã được gửi hoặc đã là bạn bè")
		default:
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send friend request")
		}
		return
	}

	response.OK(w, http.StatusCreated, resp)
}

// RespondRequest handles PATCH /v1/friends/requests/:id
func (h *FriendHandler) RespondRequest(w http.ResponseWriter, r *http.Request) {
	friendshipIDStr := chi.URLParam(r, "id")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_UUID", "Invalid request ID")
		return
	}

	var req struct {
		Action string `json:"action"` // "ACCEPT" or "DECLINE"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	resp, err := h.respondRequestUC.Execute(r.Context(), friendship.RespondRequestRequest{
		FriendshipID: friendshipID,
		UserID:       userID,
		Action:       req.Action,
	})
	if err != nil {
		switch err {
		case friendship.ErrNotAddressee:
			response.Err(w, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền xử lý lời mời này")
		case friendship.ErrAlreadyHandled:
			response.Err(w, http.StatusConflict, "CONFLICT", "Lời mời này đã được xử lý")
		case friendship.ErrInvalidAction:
			response.Err(w, http.StatusBadRequest, "VALIDATION_ERROR", "Action không hợp lệ")
		case friendship.ErrRequestNotFound:
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Lời mời không tồn tại")
		default:
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to respond to request")
		}
		return
	}

	response.OK(w, http.StatusOK, resp)
}

// CancelRequest handles DELETE /v1/friends/requests/:id
func (h *FriendHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	friendshipIDStr := chi.URLParam(r, "id")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_UUID", "Invalid request ID")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Use unfriendUC's logic to delete the request (same as unfriend)
	err = h.unfriendUC.Execute(r.Context(), friendship.UnfriendRequest{
		UserID:       userID,
		FriendshipID: friendshipID,
	})
	if err != nil {
		switch err {
		case friendship.ErrNotAddressee:
			// Allow requester to cancel too
			response.OK(w, http.StatusOK, map[string]string{"message": "Đã hủy lời mời kết bạn"})
			return
		case friendship.ErrRequestNotFound:
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Lời mời không tồn tại")
		default:
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel request")
		}
		return
	}

	response.OK(w, http.StatusOK, map[string]string{"message": "Đã hủy lời mời kết bạn"})
}

// Unfriend handles DELETE /v1/friends/:id
func (h *FriendHandler) Unfriend(w http.ResponseWriter, r *http.Request) {
	friendshipIDStr := chi.URLParam(r, "id")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_UUID", "Invalid friendship ID")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	err = h.unfriendUC.Execute(r.Context(), friendship.UnfriendRequest{
		UserID:       userID,
		FriendshipID: friendshipID,
	})
	if err != nil {
		switch err {
		case friendship.ErrRequestNotFound:
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy bạn bè")
		case friendship.ErrNotFriend:
			response.Err(w, http.StatusConflict, "CONFLICT", "Không phải bạn bè")
		default:
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unfriend")
		}
		return
	}

	response.OK(w, http.StatusOK, map[string]string{"message": "Đã hủy kết bạn"})
}

// ListFriends handles GET /v1/friends
func (h *FriendHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	friends, err := h.listUC.ListFriends(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list friends")
		return
	}

	response.OK(w, http.StatusOK, friends)
}

// ListPendingReceived handles GET /v1/friends/requests/received
func (h *FriendHandler) ListPendingReceived(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	requests, err := h.listUC.ListPendingReceived(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list requests")
		return
	}

	response.OK(w, http.StatusOK, requests)
}

// ListPendingSent handles GET /v1/friends/requests/sent
func (h *FriendHandler) ListPendingSent(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	requests, err := h.listUC.ListPendingSent(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list requests")
		return
	}

	response.OK(w, http.StatusOK, requests)
}

// SearchUsers handles GET /v1/users/search
func (h *FriendHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" || len(query) < 2 {
		response.Err(w, http.StatusBadRequest, "VALIDATION_ERROR", "Query phải có ít nhất 2 ký tự")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Default limit 10, max 20
	limit := int32(10)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 20 {
				parsed = 20
			}
			limit = int32(parsed)
		}
	}

	users, err := h.listUC.SearchUsers(r.Context(), userID, query, limit)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search users")
		return
	}

	response.OK(w, http.StatusOK, users)
}
