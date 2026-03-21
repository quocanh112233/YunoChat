package message

import "errors"

var (
	ErrMessageNotFound     = errors.New("message not found")
	ErrCannotEditMessage   = errors.New("cannot edit this message")
	ErrCannotDeleteMessage = errors.New("cannot delete this message")
)
