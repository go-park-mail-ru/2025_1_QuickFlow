package validation_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/feedback_service/utils/validation"
	"quickflow/shared/models"
)

func TestValidateFeedback(t *testing.T) {
	tests := []struct {
		name     string
		feedback *models.Feedback
		expected error
	}{
		{
			name: "Valid feedback",
			feedback: &models.Feedback{
				Id:           uuid.New(),
				Rating:       4,
				RespondentId: uuid.New(),
				Text:         "This is valid feedback.",
				Type:         models.FeedbackPost,
			},
			expected: nil,
		},
		{
			name: "Invalid respondent (nil UUID)",
			feedback: &models.Feedback{
				Id:           uuid.New(),
				Rating:       4,
				RespondentId: uuid.Nil,
				Text:         "This is valid feedback.",
				Type:         models.FeedbackPost,
			},
			expected: validation.ErrRespondent,
		},
		{
			name: "Invalid rating (too low)",
			feedback: &models.Feedback{
				Id:           uuid.New(),
				Rating:       -1,
				RespondentId: uuid.New(),
				Text:         "This is valid feedback.",
				Type:         models.FeedbackPost,
			},
			expected: validation.ErrRating,
		},
		{
			name: "Invalid rating (too high)",
			feedback: &models.Feedback{
				Id:           uuid.New(),
				Rating:       6,
				RespondentId: uuid.New(),
				Text:         "This is valid feedback.",
				Type:         models.FeedbackPost,
			},
			expected: validation.ErrRating,
		},
		{
			name: "Text too long",
			feedback: &models.Feedback{
				Id:           uuid.New(),
				Rating:       4,
				RespondentId: uuid.New(),
				Text:         "a" + string(make([]byte, 201)), // Text longer than 200 characters
				Type:         models.FeedbackPost,
			},
			expected: validation.ErrTextTooLong,
		},
		{
			name:     "Nil feedback",
			feedback: nil,
			expected: errors.New("invalid Feedback"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateFeedback(tt.feedback)
			if tt.expected != nil {
				assert.EqualError(t, err, tt.expected.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
