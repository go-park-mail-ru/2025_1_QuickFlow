package errors

import (
	"errors"
)

// Error messages for user service
var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrUserValidation = errors.New("user validation error")
)

// Error messages for profile service
var (
	ErrInvalidProfileInfo = errors.New("invalid profile info")
	ErrInvalidUserId      = errors.New("invalid user id")
	ErrProfileValidation  = errors.New("profile validation error")
)
