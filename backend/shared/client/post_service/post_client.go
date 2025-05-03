package post_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

type PostServiceClient struct {
	client pb.PostServiceClient
}

func NewPostServiceClient(conn *grpc.ClientConn) *PostServiceClient {
	return &PostServiceClient{
		client: pb.NewPostServiceClient(conn),
	}
}

func (c *PostServiceClient) AddPost(ctx context.Context, post models.Post) (*models.Post, error) {
	logger.Info(ctx, "Sending request to add post: %v", post)
	resp, err := c.client.AddPost(ctx, &pb.AddPostRequest{
		Post: ModelPostToProto(&post),
	})
	if err != nil {
		logger.Error(ctx, "Failed to add post: %v", err)
		return nil, err
	}
	return ProtoPostToModel(resp.Post)
}

func (c *PostServiceClient) DeletePost(ctx context.Context, userId, postId uuid.UUID) error {
	logger.Info(ctx, "Sending request to delete post: %v", postId)
	_, err := c.client.DeletePost(ctx, &pb.DeletePostRequest{
		PostId: postId.String(),
		UserId: userId.String(),
	})
	return err
}

func (c *PostServiceClient) UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error) {
	logger.Info(ctx, "Sending request to update post: %v", update)
	resp, err := c.client.UpdatePost(ctx, &pb.UpdatePostRequest{
		Post:   ModelPostUpdateToProto(&update),
		UserId: userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to update post: %v", err)
		return nil, err
	}
	return ProtoPostToModel(resp.Post)
}

func (c *PostServiceClient) FetchFeed(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error) {
	logger.Info(ctx, "Sending request to fetch feed: %v", numPosts)
	resp, err := c.client.FetchFeed(ctx, &pb.FetchFeedRequest{
		NumPosts:  int32(numPosts),
		Timestamp: ToTimestamp(timestamp),
		UserId:    userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to fetch feed: %v", err)
		return nil, err
	}
	return convertProtoPosts(resp.Posts)
}

func (c *PostServiceClient) FetchRecommendations(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error) {
	logger.Info(ctx, "Sending request to fetch recommendations: %v", numPosts)
	resp, err := c.client.FetchRecommendations(ctx, &pb.FetchRecommendationsRequest{
		NumPosts:  int32(numPosts),
		Timestamp: ToTimestamp(timestamp),
		UserId:    userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to fetch recommendations: %v", err)
		return nil, err
	}
	return convertProtoPosts(resp.Posts)
}

func (c *PostServiceClient) FetchUserPosts(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	logger.Info(ctx, "Sending request to fetch user posts: %v", numPosts)
	resp, err := c.client.FetchUserPosts(ctx, &pb.FetchUserPostsRequest{
		UserId:    userId.String(),
		NumPosts:  int32(numPosts),
		Timestamp: ToTimestamp(timestamp),
	})
	if err != nil {
		logger.Error(ctx, "Failed to fetch user posts: %v", err)
		return nil, err
	}
	return convertProtoPosts(resp.Posts)
}

func (c *PostServiceClient) LikePost(ctx context.Context, postId, userId uuid.UUID) error {
	logger.Info(ctx, "Sending request to like post: %v", postId)
	_, err := c.client.LikePost(ctx, &pb.LikePostRequest{
		PostId: postId.String(),
		UserId: userId.String(),
	})
	return err
}

func (c *PostServiceClient) UnlikePost(ctx context.Context, postId, userId uuid.UUID) error {
	logger.Info(ctx, "Sending request to unlike post: %v", postId)
	_, err := c.client.UnlikePost(ctx, &pb.UnlikePostRequest{
		PostId: postId.String(),
		UserId: userId.String(),
	})
	return err
}
