package errors

import (
	"errors"
)

// Error messages for user service
var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrUserValidation = errors.New("user validation error")
	ErrInvalidUserId  = errors.New("invalid user id")
)

// Error messages for profile service
var (
	ErrInvalidProfileInfo = errors.New("invalid profile info")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrProfileValidation  = errors.New("profile validation error")
)
