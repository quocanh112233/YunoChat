package message

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, m *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	ListByConversation(ctx context.Context, conversationID uuid.UUID, cursorCreatedAt time.Time, cursorID uuid.UUID, limit int) ([]*Message, error)
	Update(ctx context.Context, m *Message) error
	SoftDelete(ctx context.Context, id uuid.UUID, senderID uuid.UUID) error
	UpdateStatusForConversation(ctx context.Context, conversationID uuid.UUID, recipientID uuid.UUID, status MessageStatus, previousStatuses []MessageStatus) error
}

type AttachmentRepository interface {
	Create(ctx context.Context, a *Attachment) error
	FindByMessageID(ctx context.Context, messageID uuid.UUID) (*Attachment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
