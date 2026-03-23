package conversation

import (
	"context"
	"errors"
	"time"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMinParticipants = errors.New("group requires at least 3 participants")
	ErrInvalidName     = errors.New("group name is required")
)

// CreateGroupUseCase handles creating a new group conversation
type CreateGroupUseCase struct {
	convRepo postgres.ConversationRepository
	pool     *pgxpool.Pool
}

// NewCreateGroupUseCase creates a new use case
func NewCreateGroupUseCase(convRepo postgres.ConversationRepository, pool *pgxpool.Pool) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		convRepo: convRepo,
		pool:     pool,
	}
}

// CreateGroupRequest represents the request parameters
type CreateGroupRequest struct {
	CreatorID      string
	Name           string
	ParticipantIDs []string
}

// CreateGroupResponse represents the response
type CreateGroupResponse struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Name           string            `json:"name"`
	AvatarURL      string            `json:"avatar_url,omitempty"`
	LastActivityAt time.Time         `json:"last_activity_at"`
	Participants   []ParticipantInfo `json:"participants"`
	CreatedAt      time.Time         `json:"created_at"`
}

// Execute runs the use case
func (uc *CreateGroupUseCase) Execute(ctx context.Context, req CreateGroupRequest) (*CreateGroupResponse, error) {
	// Validate group name
	if req.Name == "" {
		return nil, ErrInvalidName
	}

	// Validate minimum participants (creator + 2 others = 3 total)
	if len(req.ParticipantIDs) < 2 {
		return nil, ErrMinParticipants
	}

	creatorID, err := parseUUID(req.CreatorID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create conversation
	convID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	conv, err := uc.convRepo.CreateConversation(ctx, tx, convID, "GROUP", req.Name)
	if err != nil {
		return nil, err
	}

	// Add creator as ADMIN
	creatorParticipantID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = uc.convRepo.CreateParticipant(ctx, tx, creatorParticipantID, convID, creatorID, "ADMIN")
	if err != nil {
		return nil, err
	}

	// Add other participants as MEMBER
	participants := []ParticipantInfo{
		{
			ID:     uuid.UUID(creatorParticipantID.Bytes).String(),
			UserID: req.CreatorID,
			Role:   "ADMIN",
		},
	}

	for _, participantIDStr := range req.ParticipantIDs {
		participantID, err := parseUUID(participantIDStr)
		if err != nil {
			return nil, err
		}

		participantUUID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		participant, err := uc.convRepo.CreateParticipant(ctx, tx, participantUUID, convID, participantID, "MEMBER")
		if err != nil {
			return nil, err
		}

		participants = append(participants, ParticipantInfo{
			ID:     uuid.UUID(participant.ID.Bytes).String(),
			UserID: participantIDStr,
			Role:   "MEMBER",
		})
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateGroupResponse{
		ID:             uuid.UUID(conv.ID.Bytes).String(),
		Type:           conv.Type,
		Name:           conv.Name.String,
		LastActivityAt: conv.LastActivityAt.Time,
		Participants:   participants,
		CreatedAt:      conv.CreatedAt.Time,
	}, nil
}
