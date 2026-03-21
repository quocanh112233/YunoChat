package friendship

import "errors"

var (
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrAlreadyFriends     = errors.New("users are already friends")
	ErrRequestPending     = errors.New("friend request is already pending")
	ErrCannotAddSelf      = errors.New("cannot add yourself as a friend")
)
