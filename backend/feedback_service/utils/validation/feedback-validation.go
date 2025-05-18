package validation

import (
	"errors"

	"github.com/google/uuid"

	feedback_errors "quickflow/feedback_service/internal/errors"
	"quickflow/shared/models"
)

func ValidateFeedback(feedback *models.Feedback) error {
	if feedback == nil {
		return errors.New("invalid Feedback")
	}
	if feedback.RespondentId == uuid.Nil {
		return feedback_errors.ErrRespondent
	}

	if feedback.Rating < 0 || (feedback.Rating > 5) {
		return feedback_errors.ErrRating
	}

	// TODO
	if len(feedback.Text) > 200 {
		return feedback_errors.ErrTextTooLong
	}

	return nil
}
