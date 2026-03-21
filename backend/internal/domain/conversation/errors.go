package conversation

import "errors"

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotAParticipant      = errors.New("user is not a participant of this conversation")
	ErrActionNotAllowed     = errors.New("action not allowed for this role")
	ErrParticipantExists    = errors.New("user is already a participant")
	ErrInvalidGroupData     = errors.New("invalid group conversation data")
)
