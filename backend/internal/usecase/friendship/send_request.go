package friendship

import (
	"context"
	"errors"

	"backend/internal/domain/user"
	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
}

// NewSendRequestUseCase creates a new SendRequestUseCase
func NewSendRequestUseCase(
	friendshipRepo postgres.FriendshipRepository,
	userRepo user.UserRepository,
	notificationRepo postgres.NotificationRepository,
) *SendRequestUseCase {
	return &SendRequestUseCase{
		friendshipRepo:   friendshipRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
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

	return &SendRequestResponse{
		FriendshipID: pgToUUID(friendship.ID),
		Status:       friendship.Status,
		CreatedAt:    friendship.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// Helper to convert pgtype.UUID to uuid.UUID
func pgToUUID(p pgtype.UUID) uuid.UUID {
	return p.Bytes
}
