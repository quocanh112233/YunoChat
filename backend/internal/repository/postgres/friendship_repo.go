package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/repository/sqlc"
)

// FriendshipRepository defines friendship operations
type FriendshipRepository interface {
	Create(ctx context.Context, requesterID, addresseeID pgtype.UUID) (sqlc.Friendship, error)
	GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error)
	GetBetweenUsers(ctx context.Context, userA, userB pgtype.UUID) (sqlc.Friendship, error)
	ListFriends(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsByUserRow, error)
	ListPendingReceived(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListPendingRequestsReceivedRow, error)
	ListPendingSent(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListPendingRequestsSentRow, error)
	UpdateStatus(ctx context.Context, id pgtype.UUID, status string) (sqlc.Friendship, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	FindDMConversation(ctx context.Context, userA, userB pgtype.UUID) (pgtype.UUID, error)
	SearchUsersWithRelationship(ctx context.Context, currentUserID pgtype.UUID, query string, limit int32) ([]sqlc.SearchUsersWithRelationshipRow, error)
}

// friendshipRepository implements FriendshipRepository using SQLC
type friendshipRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// NewFriendshipRepository creates a new FriendshipRepository
func NewFriendshipRepository(pool *pgxpool.Pool, queries *sqlc.Queries) FriendshipRepository {
	return &friendshipRepository{
		queries: queries,
		pool:    pool,
	}
}

func (r *friendshipRepository) Create(ctx context.Context, requesterID, addresseeID pgtype.UUID) (sqlc.Friendship, error) {
	return r.queries.CreateFriendship(ctx, sqlc.CreateFriendshipParams{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
	})
}

func (r *friendshipRepository) GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error) {
	return r.queries.GetFriendshipByID(ctx, id)
}

func (r *friendshipRepository) GetBetweenUsers(ctx context.Context, userA, userB pgtype.UUID) (sqlc.Friendship, error) {
	return r.queries.GetFriendshipBetweenUsers(ctx, sqlc.GetFriendshipBetweenUsersParams{
		RequesterID: userA,
		AddresseeID: userB,
	})
}

func (r *friendshipRepository) ListFriends(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsByUserRow, error) {
	return r.queries.ListFriendsByUser(ctx, userID)
}

func (r *friendshipRepository) ListPendingReceived(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListPendingRequestsReceivedRow, error) {
	return r.queries.ListPendingRequestsReceived(ctx, userID)
}

func (r *friendshipRepository) ListPendingSent(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListPendingRequestsSentRow, error) {
	return r.queries.ListPendingRequestsSent(ctx, userID)
}

func (r *friendshipRepository) UpdateStatus(ctx context.Context, id pgtype.UUID, status string) (sqlc.Friendship, error) {
	return r.queries.UpdateFriendshipStatus(ctx, sqlc.UpdateFriendshipStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *friendshipRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteFriendship(ctx, id)
}

func (r *friendshipRepository) FindDMConversation(ctx context.Context, userA, userB pgtype.UUID) (pgtype.UUID, error) {
	convID, err := r.queries.FindDMConversationBetweenUsers(ctx, sqlc.FindDMConversationBetweenUsersParams{
		UserID:   userA,
		UserID_2: userB,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	return convID, nil
}

func (r *friendshipRepository) SearchUsersWithRelationship(ctx context.Context, currentUserID pgtype.UUID, query string, limit int32) ([]sqlc.SearchUsersWithRelationshipRow, error) {
	return r.queries.SearchUsersWithRelationship(ctx, sqlc.SearchUsersWithRelationshipParams{
		RequesterID: currentUserID,
		Column2:     pgtype.Text{String: query, Valid: true},
		Limit:       limit,
	})
}
