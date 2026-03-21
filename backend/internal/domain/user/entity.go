package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Email              string    `json:"email" db:"email"`
	Username           string    `json:"username" db:"username"`
	PasswordHash       string    `json:"-" db:"password_hash"`
	DisplayName        string    `json:"display_name" db:"display_name"`
	Bio                *string   `json:"bio" db:"bio"`
	AvatarURL          *string   `json:"avatar_url" db:"avatar_url"`
	AvatarCloudinaryID *string   `json:"avatar_cloudinary_id" db:"avatar_cloudinary_id"`
	Status             string    `json:"status" db:"status"` // "ONLINE" | "OFFLINE"
	LastSeenAt         time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash  string    `json:"-" db:"token_hash"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	IsRevoked  bool      `json:"is_revoked" db:"is_revoked"`
	DeviceInfo *string   `json:"device_info" db:"device_info"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
