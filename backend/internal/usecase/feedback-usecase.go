package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"quickflow/internal/models"
	"quickflow/utils/validation"
	"time"
)

var (
	ErrRespondent  = validation.ErrRespondent
	ErrRating      = validation.ErrRating
	ErrTextTooLong = validation.ErrTextTooLong
)

type FeedbackRepository interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAverageRatingType(ctx context.Context, feedbackType models.FeedbackType) (float64, error)
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
	GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error)
}

type FeedbackService struct {
	feedbackRepo FeedbackRepository
}

func NewFeedBackService(feedbackRepo FeedbackRepository) *FeedbackService {
	return &FeedbackService{
		feedbackRepo: feedbackRepo,
	}
}

func (s *FeedbackService) SaveFeedback(ctx context.Context, feedback *models.Feedback) error {
	// validate feedback
	if err := validation.ValidateFeedback(feedback); err != nil {
		return err
	}

	return s.feedbackRepo.SaveFeedback(ctx, feedback)
}

func (s *FeedbackService) GetAllFeedback(ctx context.Context, ts time.Time, count int) (map[models.FeedbackType][]models.Feedback, error) {
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

func (s *FeedbackService) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	feedback, err := s.feedbackRepo.GetAllFeedbackType(ctx, feedbackType, ts, count)
	if err != nil {
		return nil, fmt.Errorf("get all feedback type: %w", err)
	}
	return feedback, nil
}

func (s *FeedbackService) GetAverageRatings(ctx context.Context) (map[models.FeedbackType]float64, error) {
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

func (s *FeedbackService) GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumMessagesSent(ctx, userId)
}

func (s *FeedbackService) GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumPostsCreated(ctx, userId)
}

func (s *FeedbackService) GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error) {
	return s.feedbackRepo.GetNumProfileChanges(ctx, userId)
}
