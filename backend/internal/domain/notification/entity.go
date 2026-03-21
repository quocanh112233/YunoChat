package notification

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	TypeFriendRequest  NotificationType = "FRIEND_REQUEST"
	TypeFriendAccepted NotificationType = "FRIEND_ACCEPTED"
	TypeGroupAdded     NotificationType = "GROUP_ADDED"
)

type ReferenceType string

const (
	RefFriendship   ReferenceType = "friendship"
	RefConversation ReferenceType = "conversation"
)

type Notification struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	RecipientID   uuid.UUID        `json:"recipient_id" db:"recipient_id"`
	ActorID       uuid.UUID        `json:"actor_id" db:"actor_id"`
	Type          NotificationType `json:"type" db:"type"`
	ReferenceID   uuid.UUID        `json:"reference_id" db:"reference_id"`
	ReferenceType ReferenceType    `json:"reference_type" db:"reference_type"`
	IsRead        bool             `json:"is_read" db:"is_read"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	ReadAt        *time.Time       `json:"read_at" db:"read_at"`
}
