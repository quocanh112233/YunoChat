package conversation

import (
	"context"
	"errors"

	"backend/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
)

// GetConversationUseCase handles getting details of a single conversation
type GetConversationUseCase struct {
	convRepo postgres.ConversationRepository
}

// NewGetConversationUseCase creates a new use case
func NewGetConversationUseCase(convRepo postgres.ConversationRepository) *GetConversationUseCase {
	return &GetConversationUseCase{convRepo: convRepo}
}

// GetConversationRequest represents the request parameters
type GetConversationRequest struct {
	UserID         string
	ConversationID string
}

// Execute runs the use case
func (uc *GetConversationUseCase) Execute(ctx context.Context, req GetConversationRequest) (*ConversationWithDetails, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	convID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		return nil, errors.New("invalid conversation id")
	}

	userPgID := pgtype.UUID{Bytes: userID, Valid: true}
	convPgID := pgtype.UUID{Bytes: convID, Valid: true}

	// Fetch conversation details from repository
	row, err := uc.convRepo.GetConversationDetails(ctx, userPgID, convPgID)
	if err != nil {
		return nil, ErrConversationNotFound
	}

	// Fetch participants
	participantRows, err := uc.convRepo.GetConversationParticipants(ctx, convPgID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	conv := ConversationWithDetails{
		ID:             uuid.UUID(row.ID.Bytes).String(),
		Type:           row.Type,
		LastActivityAt: row.LastActivityAt.Time,
		UnreadCount:    row.UnreadCount,
	}

	if row.Name.Valid {
		conv.Name = row.Name.String
	}
	if row.AvatarUrl.Valid {
		conv.AvatarURL = row.AvatarUrl.String
	}

	// Add friendship status if it's a DM
	if conv.Type == "DM" {
		if status, ok := row.FriendshipStatus.(string); ok {
			conv.FriendshipStatus = &status
		}
	}

	// Convert participants
	conv.Participants = make([]ParticipantInfo, len(participantRows))
	for i, p := range participantRows {
		conv.Participants[i] = ParticipantInfo{
			ID:          uuid.UUID(p.ID.Bytes).String(),
			UserID:      uuid.UUID(p.UserID.Bytes).String(),
			Username:    p.Username,
			DisplayName: p.DisplayName,
			Status:      p.Status,
			Role:        p.Role,
			JoinedAt:    p.JoinedAt.Time,
		}
		if p.AvatarUrl.Valid {
			conv.Participants[i].AvatarURL = p.AvatarUrl.String
		}
	}

	return &conv, nil
}
