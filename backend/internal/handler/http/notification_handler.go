package http

import (
	"net/http"
	"strconv"

	"backend/internal/pkg/response"
	"backend/internal/usecase/notification"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	listUC     *notification.ListUseCase
	markReadUC *notification.MarkReadUseCase
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/unread-count", h.GetUnreadCount)
		r.Patch("/read-all", h.MarkAllRead)
		r.Patch("/{id}/read", h.MarkRead)
	})
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(
	listUC *notification.ListUseCase,
	markReadUC *notification.MarkReadUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		listUC:     listUC,
		markReadUC: markReadUC,
	}
}

// List handles GET /v1/notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Parse limit and offset
	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 50 {
				parsed = 50
			}
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	resp, err := h.listUC.Execute(r.Context(), notification.ListRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list notifications")
		return
	}

	response.OKWithMeta(w, http.StatusOK, resp.Notifications, &response.Meta{
		HasMore: resp.Meta.HasMore,
	})
}

// GetUnreadCount handles GET /v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Just use the list use case to get unread count
	resp, err := h.listUC.Execute(r.Context(), notification.ListRequest{
		UserID: userID,
		Limit:  0,
		Offset: 0,
	})
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get unread count")
		return
	}

	response.OK(w, http.StatusOK, map[string]int64{"count": resp.Meta.UnreadCount})
}

// MarkRead handles PATCH /v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_UUID", "Invalid notification ID")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	resp, err := h.markReadUC.Execute(r.Context(), notification.MarkReadRequest{
		NotificationID: notificationID,
		UserID:         userID,
	})
	if err != nil {
		switch err {
		case notification.ErrNotificationNotFound:
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Thông báo không tồn tại")
		case notification.ErrNotRecipient:
			response.Err(w, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền đánh dấu thông báo này")
		default:
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read")
		}
		return
	}

	response.OK(w, http.StatusOK, resp)
}

// MarkAllRead handles PATCH /v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	resp, err := h.markReadUC.ExecuteAll(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all notifications as read")
		return
	}

	response.OK(w, http.StatusOK, resp)
}
