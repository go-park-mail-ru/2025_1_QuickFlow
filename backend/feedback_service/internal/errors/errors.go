package errors

import (
	"errors"
)

var (
	ErrRespondent  = errors.New("invalid respondent")
	ErrRating      = errors.New("invalid rating")
	ErrTextTooLong = errors.New("text too long")
	ErrNotFound    = errors.New("not found")
)
