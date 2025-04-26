package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"quickflow/feedback-microservice/feedback"
)

type Gateway struct {
	feedbackClient feedback.FeedbackServiceClient
}

func NewGateway(connectAddr string) (*Gateway, error) {
	grpcConn, err := grpc.NewClient(
		connectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	feedbackClient := feedback.NewFeedbackServiceClient(grpcConn)

	return &Gateway{feedbackClient: feedbackClient}, nil
}

func (gw *Gateway) SaveFeedback(ctx context.Context, session string) error {
	saveFeedback, err := gw.feedbackClient.SaveFeedback(ctx)
	if err != nil {
		return err
	}
}
