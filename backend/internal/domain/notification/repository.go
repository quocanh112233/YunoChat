package notification

import (
	"context"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*Notification, error)
	CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id, recipientID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error
}
