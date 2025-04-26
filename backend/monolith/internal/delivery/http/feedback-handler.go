package http

import (
	"context"
	"net/http"
	"quickflow/monolith/internal/models"
	"time"
)

type FeedbackUseCase interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedback(ctx context.Context, ts time.Time, count int) (map[models.FeedbackType][]models.Feedback, error)
	GetAverageRatings(ctx context.Context) (map[models.FeedbackType]float64, error)
}

type FeedbackHandler struct {
	feedbackUseCase FeedbackUseCase
	authService     AuthUseCase
}

func NewFeedbackService(feedbackUseCase FeedbackUseCase, authUseCase AuthUseCase) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackUseCase: feedbackUseCase,
		authService:     authUseCase,
	}
}

func (f *FeedbackHandler) GetFeed(w http.ResponseWriter, r *http.Request) {

}
