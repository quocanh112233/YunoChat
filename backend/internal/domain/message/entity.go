package message

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	TypeText       MessageType = "TEXT"
	TypeAttachment MessageType = "ATTACHMENT"
)

type MessageStatus string

const (
	StatusSent      MessageStatus = "SENT"
	StatusDelivered MessageStatus = "DELIVERED"
	StatusRead      MessageStatus = "READ"
)

type Message struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	ConversationID uuid.UUID     `json:"conversation_id" db:"conversation_id"`
	SenderID       uuid.UUID     `json:"sender_id" db:"sender_id"`
	Body           *string       `json:"body" db:"body"`
	Type           MessageType   `json:"type" db:"type"`
	Status         MessageStatus `json:"status" db:"status"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time    `json:"deleted_at" db:"deleted_at"`
}

type StorageType string

const (
	StorageCloudinary StorageType = "CLOUDINARY"
	StorageR2         StorageType = "R2"
)

type FileType string

const (
	FileImage FileType = "IMAGE"
	FileVideo FileType = "VIDEO"
	FileFile  FileType = "FILE"
)

type Attachment struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	MessageID    uuid.UUID   `json:"message_id" db:"message_id"`
	StorageType  StorageType `json:"storage_type" db:"storage_type"`
	FileType     FileType    `json:"file_type" db:"file_type"`
	URL          string      `json:"url" db:"url"`
	ThumbnailURL *string     `json:"thumbnail_url" db:"thumbnail_url"`
	OriginalName string      `json:"original_name" db:"original_name"`
	MimeType     string      `json:"mime_type" db:"mime_type"`
	SizeBytes    int64       `json:"size_bytes" db:"size_bytes"`
	Width        *int        `json:"width" db:"width"`
	Height       *int        `json:"height" db:"height"`
	DurationSecs *int        `json:"duration_secs" db:"duration_secs"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
}
