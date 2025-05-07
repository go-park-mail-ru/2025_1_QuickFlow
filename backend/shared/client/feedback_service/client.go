package feedback_service

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func (c *Client) GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error) {
	logger.Info(ctx, "Trying to get amount of messages sent for user: %s", userId.String())
	resp, err := c.client.GetNumMessagesSent(ctx, &pb.GetNumMessagesSentRequest{
		UserId: userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get number of messages sent: %v", err)
		return 0, err
	}
	return resp.NumMessagesSent, nil
}

func (c *Client) GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error) {
	logger.Info(ctx, "Trying to get amount of posts created for user: %s", userId.String())
	resp, err := c.client.GetNumPostsCreated(ctx, &pb.GetNumPostsCreatedRequest{
		UserId: userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get number of posts created: %v", err)
		return 0, err
	}
	return resp.NumPostsCreated, nil
}

func (c *Client) GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error) {
	logger.Info(ctx, "Trying to get amount of profile changes for user: %s", userId.String())
	resp, err := c.client.GetNumProfileChanges(ctx, &pb.GetNumProfileChangesRequest{
		UserId: userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get number of profile changes: %v", err)
		return 0, err
	}
	return resp.NumProfileChanges, nil
}
