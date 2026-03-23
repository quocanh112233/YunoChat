package friendship

import (
	"context"
	"errors"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFriend = errors.New("không phải bạn bè")
)

// UnfriendUseCase handles unfriending
type UnfriendUseCase struct {
	friendshipRepo postgres.FriendshipRepository
}

// NewUnfriendUseCase creates a new UnfriendUseCase
func NewUnfriendUseCase(friendshipRepo postgres.FriendshipRepository) *UnfriendUseCase {
	return &UnfriendUseCase{
		friendshipRepo: friendshipRepo,
	}
}

// UnfriendRequest represents the request to unfriend
type UnfriendRequest struct {
	UserID       uuid.UUID // Current user
	FriendshipID uuid.UUID
}

// Execute unfriends a user
func (uc *UnfriendUseCase) Execute(ctx context.Context, req UnfriendRequest) error {
	friendshipPgID := pgtype.UUID{Bytes: req.FriendshipID, Valid: true}
	userPgID := pgtype.UUID{Bytes: req.UserID, Valid: true}

	// Get friendship
	friendship, err := uc.friendshipRepo.GetByID(ctx, friendshipPgID)
	if err != nil {
		return ErrRequestNotFound
	}

	// Verify user is part of this friendship
	if friendship.RequesterID != userPgID && friendship.AddresseeID != userPgID {
		return ErrNotAddressee
	}

	// Verify status is ACCEPTED
	if friendship.Status != "ACCEPTED" {
		return ErrNotFriend
	}

	// Hard delete friendship
	return uc.friendshipRepo.Delete(ctx, friendshipPgID)
}
