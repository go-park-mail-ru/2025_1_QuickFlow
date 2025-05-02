package validation

import (
	"errors"

	"github.com/google/uuid"

	"quickflow/shared/models"
)

var (
	ErrRespondent  = errors.New("invalid respondent")
	ErrRating      = errors.New("invalid rating")
	ErrTextTooLong = errors.New("text is too long")
)

func ValidateFeedback(feedback *models.Feedback) error {
	if feedback == nil {
		return errors.New("invalid Feedback")
	}
	if feedback.RespondentId == uuid.Nil {
		return ErrRespondent
	}

	if feedback.Rating < 0 || (feedback.Rating > 5) {
		return ErrRating
	}

	// TODO
	if len(feedback.Text) > 200 {
		return ErrTextTooLong
	}

	return nil
}
