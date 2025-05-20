package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"

	"quickflow/post_service/internal/delivery/grpc/dto"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchRecommendations(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchUserPosts(ctx context.Context, userId uuid.UUID, requesterId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, userId uuid.UUID, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error)
	LikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error
	UnlikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error
	GetPost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (*models.Post, error)
}

type UserUseCase interface {
	GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error)
}

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	postUseCase PostUseCase
	userUseCase UserUseCase
}

func NewPostServiceServer(postUseCase PostUseCase, userUseCase UserUseCase) *PostServiceServer {
	return &PostServiceServer{
		postUseCase: postUseCase,
		userUseCase: userUseCase,
	}
}

func (p *PostServiceServer) AddPost(ctx context.Context, req *pb.AddPostRequest) (*pb.AddPostResponse, error) {
	logger.Info(ctx, "AddPost called")
	post, err := dto.ProtoPostToModel(req.Post)
	if err != nil {
		logger.Error(ctx, "Failed to convert proto to model:", err)
		return nil, err
	}

	result, err := p.postUseCase.AddPost(ctx, *post)
	if err != nil {
		logger.Error(ctx, "Failed to add post:", err)
		return nil, err
	}

	return &pb.AddPostResponse{Post: dto.ModelPostToProto(result)}, nil
}

func (p *PostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	logger.Info(ctx, "DeletePost called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := p.postUseCase.DeletePost(ctx, userId, postId); err != nil {
		logger.Error(ctx, "Failed to delete post:", err)
		return nil, err
	}
	return &pb.DeletePostResponse{Success: true}, nil
}

func (p *PostServiceServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	logger.Info(ctx, "UpdatePost called")
	update, err := dto.ProtoPostUpdateToModel(req.Post)
	if err != nil {
		logger.Error(ctx, "Failed to convert update payload:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	updatedPost, err := p.postUseCase.UpdatePost(ctx, *update, userId)
	if err != nil {
		logger.Error(ctx, "Failed to update post:", err)
		return nil, err
	}
	return &pb.UpdatePostResponse{Post: dto.ModelPostToProto(updatedPost)}, nil
}

func (p *PostServiceServer) FetchFeed(ctx context.Context, req *pb.FetchFeedRequest) (*pb.FetchFeedResponse, error) {
	logger.Info(ctx, "FetchFeed called")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	posts, err := p.postUseCase.FetchFeed(ctx, userId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		logger.Error(ctx, "Failed to fetch feed:", err)
		return nil, err
	}
	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}
	return &pb.FetchFeedResponse{Posts: protoPosts}, nil
}

func (p *PostServiceServer) FetchRecommendations(ctx context.Context, req *pb.FetchRecommendationsRequest) (*pb.FetchRecommendationsResponse, error) {
	logger.Info(ctx, "FetchRecommendations called")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	posts, err := p.postUseCase.FetchRecommendations(ctx, userId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		logger.Error(ctx, "Failed to fetch recommendations:", err)
		return nil, err
	}
	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}
	return &pb.FetchRecommendationsResponse{Posts: protoPosts}, nil
}

func (p *PostServiceServer) FetchUserPosts(ctx context.Context, req *pb.FetchUserPostsRequest) (*pb.FetchUserPostsResponse, error) {
	logger.Info(ctx, "FetchCreatorPosts called")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	requesterId, err := uuid.Parse(req.RequesterId)
	if err != nil {
		logger.Error(ctx, "Invalid requester ID:", err)
		return nil, err
	}

	posts, err := p.postUseCase.FetchUserPosts(ctx, userId, requesterId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		logger.Error(ctx, "Failed to fetch user posts:", err)
		return nil, err
	}
	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}
	return &pb.FetchUserPostsResponse{Posts: protoPosts}, nil
}

func (p *PostServiceServer) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error) {
	logger.Info(ctx, "LikePost called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := p.postUseCase.LikePost(ctx, postId, userId); err != nil {
		logger.Error(ctx, "Failed to like post:", err)
		return nil, err
	}
	return &pb.LikePostResponse{Success: true}, nil
}

func (p *PostServiceServer) UnlikePost(ctx context.Context, req *pb.UnlikePostRequest) (*pb.UnlikePostResponse, error) {
	logger.Info(ctx, "UnlikePost called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := p.postUseCase.UnlikePost(ctx, postId, userId); err != nil {
		logger.Error(ctx, "Failed to unlike post:", err)
		return nil, err
	}
	return &pb.UnlikePostResponse{Success: true}, nil
}

func (p *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	logger.Info(ctx, "GetPost called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}

	post, err := p.postUseCase.GetPost(ctx, postId, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get post:", err)
		return nil, err
	}
	return &pb.GetPostResponse{Post: dto.ModelPostToProto(post)}, nil
}
