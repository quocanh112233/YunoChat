package conversation

import (
	"time"

	"github.com/google/uuid"
)

type ConversationType string

const (
	TypeDM    ConversationType = "DM"
	TypeGroup ConversationType = "GROUP"
)

type Conversation struct {
	ID                 uuid.UUID        `json:"id" db:"id"`
	Type               ConversationType `json:"type" db:"type"`
	Name               *string          `json:"name" db:"name"`
	AvatarURL          *string          `json:"avatar_url" db:"avatar_url"`
	AvatarCloudinaryID *string          `json:"avatar_cloudinary_id" db:"avatar_cloudinary_id"`
	LastMessageID      *uuid.UUID       `json:"last_message_id" db:"last_message_id"`
	LastActivityAt     time.Time        `json:"last_activity_at" db:"last_activity_at"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

type ParticipantRole string

const (
	RoleMember ParticipantRole = "MEMBER"
	RoleAdmin  ParticipantRole = "ADMIN"
)

type ConversationParticipant struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	ConversationID     uuid.UUID       `json:"conversation_id" db:"conversation_id"`
	UserID             uuid.UUID       `json:"user_id" db:"user_id"`
	Role               ParticipantRole `json:"role" db:"role"`
	LastReadMessageID  *uuid.UUID      `json:"last_read_message_id" db:"last_read_message_id"`
	LastReadAt         *time.Time      `json:"last_read_at" db:"last_read_at"`
	JoinedAt           time.Time       `json:"joined_at" db:"joined_at"`
	LeftAt             *time.Time      `json:"left_at" db:"left_at"`
}
