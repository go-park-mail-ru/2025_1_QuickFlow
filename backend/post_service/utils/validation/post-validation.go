package validation

import (
	"errors"
	"time"
)

var (
	ErrInvalidNumPosts  = errors.New("invalid number of posts")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
)

type PostValidator struct{}

func NewPostValidator() *PostValidator {
	return &PostValidator{}
}

func (p *PostValidator) ValidateFeedParams(numPosts int, timestamp time.Time) error {
	if numPosts <= 0 {
		return ErrInvalidNumPosts
	}
	if timestamp.IsZero() || timestamp.After(time.Now()) {
		return ErrInvalidTimestamp
	}
	return nil
}
