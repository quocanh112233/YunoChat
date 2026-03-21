package user

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrDuplicateEmail      = errors.New("email already exists")
	ErrDuplicateUsername   = errors.New("username already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")
)
