package feedback_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		return err
	}

	_, err = c.client.SaveFeedback(ctx, &pb.SaveFeedbackRequest{Feedback: pbFeedback})
	return err
}

func (c *Client) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	pbType, err := FeedBackTypeToProto(feedbackType)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.GetAllFeedbackType(ctx, &pb.GetAllFeedbackTypeRequest{
		Type:  pbType,
		Ts:    timestamppb.New(ts),
		Count: int32(count),
	})
	if err != nil {
		return nil, err
	}

	var result []models.Feedback
	for _, fb := range resp.Feedback {
		modelFb, err := ProtoFeedbackToModel(fb)
		if err != nil {
			return nil, err
		}
		result = append(result, *modelFb)
	}

	return result, nil
}

func (c *Client) GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error) {
	resp, err := c.client.GetNumMessagesSent(ctx, &pb.GetNumMessagesSentRequest{
		UserId: userId.String(),
	})
	if err != nil {
		return 0, err
	}
	return resp.NumMessagesSent, nil
}

func (c *Client) GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error) {
	resp, err := c.client.GetNumPostsCreated(ctx, &pb.GetNumPostsCreatedRequest{
		UserId: userId.String(),
	})
	if err != nil {
		return 0, err
	}
	return resp.NumPostsCreated, nil
}

func (c *Client) GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error) {
	resp, err := c.client.GetNumProfileChanges(ctx, &pb.GetNumProfileChangesRequest{
		UserId: userId.String(),
	})
	if err != nil {
		return 0, err
	}
	return resp.NumProfileChanges, nil
}
