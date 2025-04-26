package models

import (
	"github.com/google/uuid"
	"time"
)

type FeedbackType string

const (
	FeedbackGeneral        = "general"
	FeedbackPost           = "post_creation"
	FeedbackMessenger      = "messenger"
	FeedbackRecommendation = "recommendation"
	FeedbackProfile        = "profile"
	FeedbackAuth           = "auth"
)

var FeedBackTypes = []FeedbackType{
	FeedbackGeneral,
	FeedbackPost,
	FeedbackRecommendation,
	FeedbackMessenger,
}

type Feedback struct {
	Id           uuid.UUID
	Rating       int
	RespondentId uuid.UUID
	Text         string
	Type         FeedbackType
	CreatedAt    time.Time
}

type AverageStat struct {
	FeedbackId    uuid.UUID
	AverageRating float32
}
