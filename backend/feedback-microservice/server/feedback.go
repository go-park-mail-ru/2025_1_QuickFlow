package server

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"

	"quickflow/feedback-microservice/feedback"
	"quickflow/monolith/internal/usecase"
)

type FeedbackManager struct {
	feedback.UnimplementedFeedbackServiceServer
	FeedbackUseCase usecase.FeedbackService
}

func NewFeedbackManager(feedbackService usecase.FeedbackService) *FeedbackManager {
	return &FeedbackManager{
		FeedbackUseCase: feedbackService,
	}
}

func (f *FeedbackManager) GetFeedbackList(ctx context.Context, in *feedback.GetFeedbackRequest) (*feedback.GetFeedbackListResponse, error) {
	resp, err := f.FeedbackUseCase.GetAllFeedback(ctx, in.Ts.AsTime(), int(in.FeedbackCount))
	if err != nil {
		return &feedback.GetFeedbackListResponse{}, err
	}

	return resp, nil
}

func (f *FeedbackManager) SaveFeedback(ctx context.Context, in *feedback.SaveFeedbackRequest) (*emptypb.Empty, error) {
	err := f.FeedbackUseCase.SaveFeedback(ctx, in)
	return nil, nil
}
