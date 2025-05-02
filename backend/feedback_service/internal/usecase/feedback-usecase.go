package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"quickflow/feedback_service/utils/validation"
	"quickflow/shared/models"
)

type FeedbackRepository interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAverageRatingType(ctx context.Context, feedbackType models.FeedbackType) (float64, error)
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
	GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error)
}

type FeedbackUseCase struct {
	feedbackRepo FeedbackRepository
}

func NewFeedBackUseCase(feedbackRepo FeedbackRepository) *FeedbackUseCase {
	return &FeedbackUseCase{
		feedbackRepo: feedbackRepo,
	}
}

func (s *FeedbackUseCase) SaveFeedback(ctx context.Context, feedback *models.Feedback) error {
	// validate feedback
	if err := validation.ValidateFeedback(feedback); err != nil {
		return err
	}

	return s.feedbackRepo.SaveFeedback(ctx, feedback)
}

func (s *FeedbackUseCase) GetAllFeedback(ctx context.Context, ts time.Time, count int) (map[models.FeedbackType][]models.Feedback, error) {
	var feedbacks = make(map[models.FeedbackType][]models.Feedback)
	feedbackTypes := models.FeedBackTypes

	for _, feedbackType := range feedbackTypes {
		feedback, err := s.feedbackRepo.GetAllFeedbackType(ctx, feedbackType, ts, count)
		if err != nil {
			return nil, fmt.Errorf("get all feedback type: %w", err)
		}
		feedbacks[feedbackType] = feedback
	}
	return feedbacks, nil
}

func (s *FeedbackUseCase) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	feedback, err := s.feedbackRepo.GetAllFeedbackType(ctx, feedbackType, ts, count)
	if err != nil {
		return nil, fmt.Errorf("get all feedback type: %w", err)
	}
	return feedback, nil
}

func (s *FeedbackUseCase) GetAverageRatings(ctx context.Context) (map[models.FeedbackType]float64, error) {
	var feedbacks = make(map[models.FeedbackType]float64)
	feedbackTypes := models.FeedBackTypes

	for _, feedbackType := range feedbackTypes {
		feedback, err := s.feedbackRepo.GetAverageRatingType(ctx, feedbackType)
		if err != nil {
			return nil, fmt.Errorf("get average rating type: %w", err)
		}
		feedbacks[feedbackType] = feedback
	}
	return feedbacks, nil
}

func (s *FeedbackUseCase) GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumMessagesSent(ctx, userId)
}

func (s *FeedbackUseCase) GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumPostsCreated(ctx, userId)
}

func (s *FeedbackUseCase) GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumProfileChanges(ctx, userId)
}
