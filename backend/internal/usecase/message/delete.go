package message

import (
	"context"
	"errors"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
)

var (
	ErrNotSender       = errors.New("only the sender can delete this message")
	ErrAlreadyDeleted  = errors.New("message has already been deleted")
	ErrMessageNotFound = errors.New("message not found")
)

// SoftDeleteUseCase handles soft deleting a message
type SoftDeleteUseCase struct {
	msgRepo postgres.MessageRepository
}

// NewSoftDeleteUseCase creates a new use case
func NewSoftDeleteUseCase(msgRepo postgres.MessageRepository) *SoftDeleteUseCase {
	return &SoftDeleteUseCase{msgRepo: msgRepo}
}

// SoftDeleteRequest represents the request parameters
type SoftDeleteRequest struct {
	UserID    string
	MessageID string
}

// SoftDeleteResponse represents the response
type SoftDeleteResponse struct {
	MessageID string `json:"message_id"`
	Deleted   bool   `json:"deleted"`
}

// Execute runs the use case
func (uc *SoftDeleteUseCase) Execute(ctx context.Context, req SoftDeleteRequest) (*SoftDeleteResponse, error) {
	senderID, err := parseUUID(req.UserID)
	if err != nil {
		return nil, err
	}

	messageID, err := parseUUID(req.MessageID)
	if err != nil {
		return nil, err
	}

	// Get message to verify sender (optional, since SQL query also checks)
	msg, err := uc.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, ErrMessageNotFound
	}

	// Verify sender
	if uuid.UUID(msg.SenderID.Bytes).String() != req.UserID {
		return nil, ErrNotSender
	}

	// Check if already deleted
	if !msg.DeletedAt.Time.IsZero() {
		return nil, ErrAlreadyDeleted
	}

	// Soft delete
	rowsAffected, err := uc.msgRepo.SoftDeleteMessage(ctx, messageID, senderID)
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNotSender
	}

	return &SoftDeleteResponse{
		MessageID: req.MessageID,
		Deleted:   true,
	}, nil
}
