package notification

import (
	"context"
	"errors"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotificationNotFound = errors.New("thông báo không tồn tại")
	ErrNotRecipient         = errors.New("bạn không có quyền đánh dấu thông báo này")
)

// MarkReadUseCase handles marking notifications as read
type MarkReadUseCase struct {
	notificationRepo postgres.NotificationRepository
}

// NewMarkReadUseCase creates a new MarkReadUseCase
func NewMarkReadUseCase(notificationRepo postgres.NotificationRepository) *MarkReadUseCase {
	return &MarkReadUseCase{
		notificationRepo: notificationRepo,
	}
}

// MarkReadRequest represents the request to mark a notification as read
type MarkReadRequest struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID // Current user (must be recipient)
}

// MarkReadResponse represents the response after marking as read
type MarkReadResponse struct {
	Message string `json:"message"`
}

// MarkAllReadResponse represents the response after marking all as read
type MarkAllReadResponse struct {
	UpdatedCount int64 `json:"updated_count"`
}

// Execute marks a single notification as read
func (uc *MarkReadUseCase) Execute(ctx context.Context, req MarkReadRequest) (*MarkReadResponse, error) {
	notifPgID := pgtype.UUID{Bytes: req.NotificationID, Valid: true}
	userPgID := pgtype.UUID{Bytes: req.UserID, Valid: true}

	// Verify notification exists and belongs to user
	notification, err := uc.notificationRepo.GetByID(ctx, notifPgID)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	if notification.RecipientID != userPgID {
		return nil, ErrNotRecipient
	}

	// Mark as read
	_, err = uc.notificationRepo.MarkRead(ctx, notifPgID, userPgID)
	if err != nil {
		return nil, err
	}

	return &MarkReadResponse{
		Message: "Đã đánh dấu đã đọc",
	}, nil
}

// ExecuteAll marks all notifications as read for a user
func (uc *MarkReadUseCase) ExecuteAll(ctx context.Context, userID uuid.UUID) (*MarkAllReadResponse, error) {
	userPgID := pgtype.UUID{Bytes: userID, Valid: true}

	// Get unread count before marking
	unreadCount, err := uc.notificationRepo.GetUnreadCount(ctx, userPgID)
	if err != nil {
		unreadCount = 0
	}

	// Mark all as read
	err = uc.notificationRepo.MarkAllRead(ctx, userPgID)
	if err != nil {
		return nil, err
	}

	return &MarkAllReadResponse{
		UpdatedCount: unreadCount,
	}, nil
}
