package friendship

import (
	"time"

	"github.com/google/uuid"
)

type FriendshipStatus string

const (
	StatusPending  FriendshipStatus = "PENDING"
	StatusAccepted FriendshipStatus = "ACCEPTED"
)

type Friendship struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	RequesterID uuid.UUID        `json:"requester_id" db:"requester_id"`
	AddresseeID uuid.UUID        `json:"addressee_id" db:"addressee_id"`
	Status      FriendshipStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}
