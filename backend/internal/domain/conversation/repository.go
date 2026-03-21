package conversation

import (
	"context"

	"github.com/google/uuid"
)

type ConversationRepository interface {
	Create(ctx context.Context, c *Conversation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindByExternalUsers(ctx context.Context, userA, userB uuid.UUID) (*Conversation, error) // For DMs
	Update(ctx context.Context, c *Conversation) error
	UpdateLastMessage(ctx context.Context, id uuid.UUID, lastMessageID uuid.UUID) error
}

type ParticipantRepository interface {
	Add(ctx context.Context, p *ConversationParticipant) error
	Remove(ctx context.Context, conversationID, userID uuid.UUID) error
	FindByConversationAndUser(ctx context.Context, conversationID, userID uuid.UUID) (*ConversationParticipant, error)
	ListByConversation(ctx context.Context, conversationID uuid.UUID) ([]*ConversationParticipant, error)
	UpdateLastRead(ctx context.Context, conversationID, userID uuid.UUID, lastMessageID uuid.UUID) error
}
