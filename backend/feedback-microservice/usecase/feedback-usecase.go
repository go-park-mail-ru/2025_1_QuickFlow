package usecase

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"quickflow/feedback-microservice/proto"
	"time"

	"quickflow/monolith/utils/validation"
)

var (
	ErrRespondent  = validation.ErrRespondent
	ErrRating      = validation.ErrRating
	ErrTextTooLong = validation.ErrTextTooLong
)

type FeedbackRepository interface {
	SaveFeedback(ctx context.Context, feedback *proto.SaveFeedbackRequest) error
	GetAverageRatingType(ctx context.Context, feedbackType models.FeedbackType) (float64, error)
	GetAllFeedback(ctx context.Context, in *proto.GetFeedbackRequest) (*proto.GetFeedbackListResponse, error)
}

type FeedbackService struct {
	feedbackRepo FeedbackRepository
}

func NewFeedBackService(feedbackRepo FeedbackRepository) *FeedbackService {
	return &FeedbackService{
		feedbackRepo: feedbackRepo,
	}
}

func (f *FeedbackService) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	feedback, err := f.feedbackRepo.GetAllFeedbackType(ctx, feedbackType, ts, count)
	if err != nil {
		return nil, fmt.Errorf("get all proto type: %w", err)
	}
	return feedback, nil
}

func (f *FeedbackService) GetFeedbackList(ctx context.Context, in *proto.GetFeedbackRequest) (*proto.GetFeedbackListResponse, error) {
	resp, err := f.feedbackRepo.GetAllFeedback(ctx, in)
	if err != nil {
		return &proto.GetFeedbackListResponse{}, err
	}

	return resp, nil
}

func (f *FeedbackService) SaveFeedback(ctx context.Context, in *proto.SaveFeedbackRequest) (*emptypb.Empty, error) {
	err := f.feedbackRepo.SaveFeedback(ctx, in)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
