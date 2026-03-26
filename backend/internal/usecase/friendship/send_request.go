package friendship

import (
	"context"
	"encoding/json"
	"errors"

	"backend/internal/domain/user"
	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSelfRequest      = errors.New("không thể gửi lời mời cho chính mình")
	ErrUserNotFound     = errors.New("người dùng không tồn tại")
	ErrDuplicateRequest = errors.New("lời mời đã được gửi hoặc đã là bạn bè")
	ErrRequestNotFound  = errors.New("lời mời không tồn tại")
)

// SendRequestUseCase handles sending friend requests
type SendRequestUseCase struct {
	friendshipRepo   postgres.FriendshipRepository
	userRepo         user.UserRepository
	notificationRepo postgres.NotificationRepository
	pool             *pgxpool.Pool
}

// NewSendRequestUseCase creates a new SendRequestUseCase
func NewSendRequestUseCase(
	friendshipRepo postgres.FriendshipRepository,
	userRepo user.UserRepository,
	notificationRepo postgres.NotificationRepository,
	pool *pgxpool.Pool,
) *SendRequestUseCase {
	return &SendRequestUseCase{
		friendshipRepo:   friendshipRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		pool:             pool,
	}
}

// SendRequestRequest represents the request to send a friend request
type SendRequestRequest struct {
	RequesterID uuid.UUID
	AddresseeID uuid.UUID
}

// SendRequestResponse represents the response after sending a friend request
type SendRequestResponse struct {
	FriendshipID uuid.UUID `json:"friendship_id"`
	Status       string    `json:"status"`
	CreatedAt    string    `json:"created_at"`
}

// Execute sends a friend request
func (uc *SendRequestUseCase) Execute(ctx context.Context, req SendRequestRequest) (*SendRequestResponse, error) {
	// Validate: cannot send to self
	if req.RequesterID == req.AddresseeID {
		return nil, ErrSelfRequest
	}

	// Check if addressee exists
	requesterPgID := pgtype.UUID{Bytes: req.RequesterID, Valid: true}
	addresseePgID := pgtype.UUID{Bytes: req.AddresseeID, Valid: true}

	_, err := uc.userRepo.FindByID(ctx, req.AddresseeID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if friendship already exists (PENDING or ACCEPTED)
	existing, err := uc.friendshipRepo.GetBetweenUsers(ctx, requesterPgID, addresseePgID)
	if err == nil && existing.ID.Valid {
		return nil, ErrDuplicateRequest
	}

	// Create friendship
	friendship, err := uc.friendshipRepo.Create(ctx, requesterPgID, addresseePgID)
	if err != nil {
		return nil, err
	}

	// Create notification for recipient
	_, _ = uc.notificationRepo.Create(
		ctx,
		addresseePgID,
		requesterPgID,
		"FRIEND_REQUEST",
		friendship.ID.String(),
		"friendship",
	)

	// Broadcast via WS
	payload := map[string]interface{}{
		"type": "notification_new",
		"data": map[string]interface{}{
			"type":         "FRIEND_REQUEST",
			"reference_id": pgToUUID(friendship.ID).String(),
			"actor_id":     req.RequesterID.String(),
		},
		"recipient_ids": []string{req.AddresseeID.String()},
	}
	jsonPayload, _ := json.Marshal(payload)
	_, _ = uc.pool.Exec(ctx, "SELECT pg_notify('chat_events', $1)", string(jsonPayload))

	return &SendRequestResponse{
		FriendshipID: pgToUUID(friendship.ID),
		Status:       friendship.Status,
		CreatedAt:    friendship.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}
