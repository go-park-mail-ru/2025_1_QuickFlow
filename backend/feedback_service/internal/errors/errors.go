package errors

import (
	"errors"

	"quickflow/feedback_service/utils/validation"
)

var (
	ErrRespondent  = validation.ErrRespondent
	ErrRating      = validation.ErrRating
	ErrTextTooLong = validation.ErrTextTooLong
	ErrNotFound    = errors.New("not found")
)
