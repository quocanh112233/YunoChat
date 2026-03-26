package postgres

import (
	"context"

	"backend/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConversationRepository defines the interface for conversation data access
type ConversationRepository interface {
	// Queries
	ListConversationsByUser(ctx context.Context, userID pgtype.UUID, cursorTime *pgtype.Timestamptz, cursorID *pgtype.UUID, limit int32) ([]sqlc.ListConversationsByUserRow, error)
	FindDMConversation(ctx context.Context, userA, userB pgtype.UUID) (pgtype.UUID, error)
	GetConversationByID(ctx context.Context, id pgtype.UUID) (sqlc.GetConversationByIDRow, error)
	GetConversationDetails(ctx context.Context, userID, conversationID pgtype.UUID) (sqlc.GetConversationDetailsRow, error)
	GetConversationParticipants(ctx context.Context, conversationID pgtype.UUID) ([]sqlc.GetConversationParticipantsRow, error)
	IsConversationMember(ctx context.Context, conversationID, userID pgtype.UUID) (bool, error)
	IsGroupAdmin(ctx context.Context, conversationID, userID pgtype.UUID) (bool, error)
	GetFriendshipStatus(ctx context.Context, userA, userB pgtype.UUID) (string, error)
	UpdateLastRead(ctx context.Context, conversationID, userID pgtype.UUID, lastMessageID pgtype.UUID) (sqlc.UpdateLastReadRow, error)
	UpdateConversation(ctx context.Context, id pgtype.UUID, name, avatarURL, avatarCloudinaryID *string) (sqlc.UpdateConversationRow, error)

	// Commands (with transaction support)
	CreateConversation(ctx context.Context, tx pgx.Tx, id pgtype.UUID, convType, name string) (sqlc.CreateConversationRow, error)
	CreateParticipant(ctx context.Context, tx pgx.Tx, id, conversationID, userID pgtype.UUID, role string) (sqlc.CreateParticipantRow, error)
	UpdateLastActivity(ctx context.Context, tx pgx.Tx, conversationID pgtype.UUID) error
	UpdateLastMessage(ctx context.Context, tx pgx.Tx, conversationID, messageID pgtype.UUID) error
}

// conversationRepository implements ConversationRepository
type conversationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(pool *pgxpool.Pool) ConversationRepository {
	return &conversationRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// ListConversationsByUser implements ConversationRepository
func (r *conversationRepository) ListConversationsByUser(ctx context.Context, userID pgtype.UUID, cursorTime *pgtype.Timestamptz, cursorID *pgtype.UUID, limit int32) ([]sqlc.ListConversationsByUserRow, error) {
	params := sqlc.ListConversationsByUserParams{
		UserID: userID,
		Limit:  limit,
	}

	if cursorTime != nil {
		params.Column2 = *cursorTime
	}
	if cursorID != nil {
		params.Column3 = *cursorID
	}

	return r.queries.ListConversationsByUser(ctx, params)
}

// FindDMConversation implements ConversationRepository
func (r *conversationRepository) FindDMConversation(ctx context.Context, userA, userB pgtype.UUID) (pgtype.UUID, error) {
	return r.queries.FindDMConversation(ctx, sqlc.FindDMConversationParams{
		UserID:   userA,
		UserID_2: userB,
	})
}

// GetConversationByID implements ConversationRepository
func (r *conversationRepository) GetConversationByID(ctx context.Context, id pgtype.UUID) (sqlc.GetConversationByIDRow, error) {
	return r.queries.GetConversationByID(ctx, id)
}

// GetConversationDetails implements ConversationRepository
func (r *conversationRepository) GetConversationDetails(ctx context.Context, userID, conversationID pgtype.UUID) (sqlc.GetConversationDetailsRow, error) {
	return r.queries.GetConversationDetails(ctx, sqlc.GetConversationDetailsParams{
		UserID: userID,
		ID:     conversationID,
	})
}

// GetConversationParticipants implements ConversationRepository
func (r *conversationRepository) GetConversationParticipants(ctx context.Context, conversationID pgtype.UUID) ([]sqlc.GetConversationParticipantsRow, error) {
	return r.queries.GetConversationParticipants(ctx, conversationID)
}

// IsConversationMember implements ConversationRepository
func (r *conversationRepository) IsConversationMember(ctx context.Context, conversationID, userID pgtype.UUID) (bool, error) {
	return r.queries.IsConversationMember(ctx, sqlc.IsConversationMemberParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
}

// IsGroupAdmin implements ConversationRepository
func (r *conversationRepository) IsGroupAdmin(ctx context.Context, conversationID, userID pgtype.UUID) (bool, error) {
	return r.queries.IsGroupAdmin(ctx, sqlc.IsGroupAdminParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
}

// GetFriendshipStatus implements ConversationRepository
func (r *conversationRepository) GetFriendshipStatus(ctx context.Context, userA, userB pgtype.UUID) (string, error) {
	return r.queries.GetFriendshipStatus(ctx, sqlc.GetFriendshipStatusParams{
		RequesterID: userA,
		AddresseeID: userB,
	})
}

// UpdateLastRead implements ConversationRepository
func (r *conversationRepository) UpdateLastRead(ctx context.Context, conversationID, userID pgtype.UUID, lastMessageID pgtype.UUID) (sqlc.UpdateLastReadRow, error) {
	return r.queries.UpdateLastRead(ctx, sqlc.UpdateLastReadParams{
		ConversationID:    conversationID,
		UserID:            userID,
		LastReadMessageID: lastMessageID,
	})
}

// UpdateConversation implements ConversationRepository
func (r *conversationRepository) UpdateConversation(ctx context.Context, id pgtype.UUID, name, avatarURL, avatarCloudinaryID *string) (sqlc.UpdateConversationRow, error) {
	params := sqlc.UpdateConversationParams{ID: id}

	if name != nil {
		params.Name = pgtype.Text{String: *name, Valid: true}
	}
	if avatarURL != nil {
		params.AvatarUrl = pgtype.Text{String: *avatarURL, Valid: true}
	}
	if avatarCloudinaryID != nil {
		params.AvatarCloudinaryID = pgtype.Text{String: *avatarCloudinaryID, Valid: true}
	}

	return r.queries.UpdateConversation(ctx, params)
}

// CreateConversation implements ConversationRepository
func (r *conversationRepository) CreateConversation(ctx context.Context, tx pgx.Tx, id pgtype.UUID, convType, name string) (sqlc.CreateConversationRow, error) {
	queries := sqlc.New(tx)

	params := sqlc.CreateConversationParams{
		ID:   id,
		Type: convType,
	}
	if name != "" {
		params.Name = pgtype.Text{String: name, Valid: true}
	}

	return queries.CreateConversation(ctx, params)
}

// CreateParticipant implements ConversationRepository
func (r *conversationRepository) CreateParticipant(ctx context.Context, tx pgx.Tx, id, conversationID, userID pgtype.UUID, role string) (sqlc.CreateParticipantRow, error) {
	queries := sqlc.New(tx)
	return queries.CreateParticipant(ctx, sqlc.CreateParticipantParams{
		ID:             id,
		ConversationID: conversationID,
		UserID:         userID,
		Role:           role,
	})
}

// UpdateLastActivity implements ConversationRepository
func (r *conversationRepository) UpdateLastActivity(ctx context.Context, tx pgx.Tx, conversationID pgtype.UUID) error {
	queries := sqlc.New(tx)
	return queries.UpdateLastActivity(ctx, conversationID)
}

// UpdateLastMessage implements ConversationRepository
func (r *conversationRepository) UpdateLastMessage(ctx context.Context, tx pgx.Tx, conversationID, messageID pgtype.UUID) error {
	queries := sqlc.New(tx)
	return queries.UpdateLastMessage(ctx, sqlc.UpdateLastMessageParams{
		ID:            conversationID,
		LastMessageID: messageID,
	})
}
