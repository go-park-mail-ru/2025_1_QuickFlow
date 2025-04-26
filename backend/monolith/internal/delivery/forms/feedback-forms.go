package forms

import (
	"errors"
	"quickflow/monolith/internal/models"
)

var InvalidTypeError = errors.New("invalid type")

const (
	FeedbackGeneral        = "general"
	FeedbackRecommendation = "recommendation"
	FeedbackPost           = "post"
	FeedbackProfile        = "profile"
	FeedbackAuth           = "auth"
	FeedbackMessenger      = "messenger"
)

type FeedbackForm struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Rating int    `json:"rating"`
}

func convertTypeToModel(Type string) (models.FeedbackType, error) {
	switch Type {
	case FeedbackGeneral:
		return models.FeedbackGeneral, nil
	case FeedbackRecommendation:
		return models.FeedbackRecommendation, nil
	case FeedbackPost:
		return models.FeedbackPost, nil
	case FeedbackProfile:
		return models.FeedbackProfile, nil
	case FeedbackAuth:
		return models.FeedbackAuth, nil
	}
	return models.FeedbackGeneral, InvalidTypeError
}

func (f *FeedbackForm) ToFeedback() (*models.Feedback, error) {
	return &models.Feedback{}
}
