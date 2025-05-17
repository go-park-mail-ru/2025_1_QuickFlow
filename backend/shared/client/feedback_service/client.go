package feedback_service

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/feedback_service"
)

type Client struct {
	client pb.FeedbackServiceClient
}

func NewFeedbackClient(conn *grpc.ClientConn) *Client {
	return &Client{
		client: pb.NewFeedbackServiceClient(conn),
	}
}

func (c *Client) SaveFeedback(ctx context.Context, feedback *models.Feedback) error {
	pbFeedback, err := ModelFeedbackToProto(feedback)
	if err != nil {
		logger.Error(ctx, "Failed to convert feedback to proto: %v", err)
		return err
	}

	logger.Info(ctx, "Saving feedback: %v", pbFeedback)
	_, err = c.client.SaveFeedback(ctx, &pb.SaveFeedbackRequest{Feedback: pbFeedback})
	return err
}

func (c *Client) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	pbType, err := FeedBackTypeToProto(feedbackType)
	if err != nil {
		logger.Error(ctx, "Failed to convert feedback type to proto: %v", err)
		return nil, err
	}

	logger.Info(ctx, "Getting all feedback of type: %v", pbType)
	resp, err := c.client.GetAllFeedbackType(ctx, &pb.GetAllFeedbackTypeRequest{
		Type:  pbType,
		Ts:    timestamppb.New(ts),
		Count: int32(count),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get all feedback type: %v", err)
		return nil, err
	}

	var result []models.Feedback
	for _, fb := range resp.Feedback {
		modelFb, err := ProtoFeedbackToModel(fb)
		if err != nil {
			logger.Error(ctx, "Failed to convert proto feedback to model: %v", err)
			return nil, err
		}
		result = append(result, *modelFb)
	}

	return result, nil
}
