package conversation

import (
	"context"
	"errors"
	"time"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

// MarkReadUseCase handles marking a conversation as read
type MarkReadUseCase struct {
	convRepo postgres.ConversationRepository
	msgRepo  postgres.MessageRepository
}

// NewMarkReadUseCase creates a new use case
func NewMarkReadUseCase(convRepo postgres.ConversationRepository, msgRepo postgres.MessageRepository) *MarkReadUseCase {
	return &MarkReadUseCase{
		convRepo: convRepo,
		msgRepo:  msgRepo,
	}
}

// MarkReadRequest represents the request parameters
type MarkReadRequest struct {
	UserID         string
	ConversationID string
	LastMessageID  string
}

// MarkReadResponse represents the response
type MarkReadResponse struct {
	LastReadMessageID string `json:"last_read_message_id"`
	ReadAt            string `json:"read_at"`
}

// Execute runs the use case
func (uc *MarkReadUseCase) Execute(ctx context.Context, req MarkReadRequest) (*MarkReadResponse, error) {
	userID, err := parseUUID(req.UserID)
	if err != nil {
		return nil, err
	}

	conversationID, err := parseUUID(req.ConversationID)
	if err != nil {
		return nil, err
	}

	// Verify user is a member
	isMember, err := uc.convRepo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	// If no message ID provided, get the latest message ID
	lastMessageID := pgtype.UUID{}
	if req.LastMessageID == "" {
		latestID, err := uc.msgRepo.GetLatestMessageID(ctx, conversationID)
		if err != nil {
			return nil, err
		}
		lastMessageID = latestID
	} else {
		id, err := uuid.Parse(req.LastMessageID)
		if err != nil {
			return nil, err
		}
		lastMessageID = pgtype.UUID{Bytes: id, Valid: true}
	}

	// Verify the message exists and belongs to this conversation
	if lastMessageID.Valid {
		msg, err := uc.msgRepo.GetMessageByID(ctx, lastMessageID)
		if err != nil {
			return nil, ErrMessageNotFound
		}
		if uuid.UUID(msg.ConversationID.Bytes).String() != req.ConversationID {
			return nil, ErrMessageNotFound
		}
	}

	// Update last read
	result, err := uc.convRepo.UpdateLastRead(ctx, conversationID, userID, lastMessageID)
	if err != nil {
		return nil, err
	}

	return &MarkReadResponse{
		LastReadMessageID: uuid.UUID(result.LastReadMessageID.Bytes).String(),
		ReadAt:            result.LastReadAt.Time.Format(time.RFC3339),
	}, nil
}
