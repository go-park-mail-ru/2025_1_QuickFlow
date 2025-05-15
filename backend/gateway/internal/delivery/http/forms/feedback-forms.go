package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
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

type FeedbackFormOut struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Rating    int    `json:"rating"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type FeedbackOutAverage struct {
	Average   float64           `json:"average"`
	Feedbacks []FeedbackFormOut `json:"feedbacks"`
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
	case FeedbackMessenger:
		return models.FeedbackMessenger, nil
	}
	return models.FeedbackGeneral, InvalidTypeError
}

func (f *FeedbackForm) ToFeedback(respondent uuid.UUID) (*models.Feedback, error) {
	modelType, err := convertTypeToModel(f.Type)
	if err != nil {
		return nil, err
	}
	feedback := &models.Feedback{
		Id:           uuid.New(),
		Rating:       f.Rating,
		Text:         f.Text,
		Type:         modelType,
		CreatedAt:    time.Now(),
		RespondentId: respondent,
	}
	return feedback, nil
}

func FromFeedBack(feedback models.Feedback, info models.PublicUserInfo) FeedbackFormOut {
	return FeedbackFormOut{
		Type:      string(feedback.Type),
		Text:      feedback.Text,
		Rating:    feedback.Rating,
		Username:  info.Username,
		Firstname: info.Firstname,
		Lastname:  info.Lastname,
	}
}

type GetFeedbackForm struct {
	Ts    time.Time           `json:"ts"`
	Count int                 `json:"count"`
	Type  models.FeedbackType `json:"type"`
}

func (f *GetFeedbackForm) GetParams(values url.Values) error {
	var (
		err      error
		numChats int64
	)

	if !values.Has("feedback_count") {
		return errors.New("chats_count parameter missing")
	}

	numChats, err = strconv.ParseInt(values.Get("feedback_count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse feedback_count")
	}

	f.Count = int(numChats)

	ts, err := time.Parse(time2.TimeStampLayout, values.Get("ts"))
	if err != nil {
		ts = time.Now()
	}
	f.Ts = ts

	f.Type = models.FeedbackType(values.Get("type"))
	return nil
}
