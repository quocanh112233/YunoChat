package friendship

import (
	"context"

	"github.com/google/uuid"
)

type FriendshipRepository interface {
	Create(ctx context.Context, f *Friendship) error
	FindByID(ctx context.Context, id uuid.UUID) (*Friendship, error)
	FindByUsers(ctx context.Context, userA, userB uuid.UUID) (*Friendship, error)
	ListFriends(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Friendship, error)
	ListPendingRequests(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Friendship, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status FriendshipStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
