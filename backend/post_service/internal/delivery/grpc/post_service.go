package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/post_service/internal/delivery/grpc/dto"
	post_errors "quickflow/post_service/internal/errors"
	"quickflow/shared/models"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchRecommendations(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchUserPosts(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, userId uuid.UUID, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error)
}

type UserUseCase interface {
	GetUserById(ctx context.Context, userId uuid.UUID) (*shared_models.User, error)
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
	post, err := dto.ProtoPostToModel(req.Post)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	result, err := p.postUseCase.AddPost(ctx, *post)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.AddPostResponse{Post: dto.ModelPostToProto(result)}, nil
}

func (p *PostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid post id: %w", post_errors.ErrInvalidUUID, err))
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	if err := p.postUseCase.DeletePost(ctx, userId, postId); err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.DeletePostResponse{Success: true}, nil
}

func (p *PostServiceServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	update, err := dto.ProtoPostUpdateToModel(req.Post)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	updatedPost, err := p.postUseCase.UpdatePost(ctx, *update, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.UpdatePostResponse{Post: dto.ModelPostToProto(updatedPost)}, nil
}

func (p *PostServiceServer) FetchFeed(ctx context.Context, req *pb.FetchFeedRequest) (*pb.FetchFeedResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	posts, err := p.postUseCase.FetchFeed(ctx, userId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}

	return &pb.FetchFeedResponse{Posts: protoPosts}, nil
}

func (p *PostServiceServer) FetchRecommendations(ctx context.Context, req *pb.FetchRecommendationsRequest) (*pb.FetchRecommendationsResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	posts, err := p.postUseCase.FetchRecommendations(ctx, userId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}

	return &pb.FetchRecommendationsResponse{Posts: protoPosts}, nil
}

func (p *PostServiceServer) FetchUserPosts(ctx context.Context, req *pb.FetchUserPostsRequest) (*pb.FetchUserPostsResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	posts, err := p.postUseCase.FetchUserPosts(ctx, userId, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = dto.ModelPostToProto(&post)
	}

	return &pb.FetchUserPostsResponse{Posts: protoPosts}, nil
}

func grpcErrorFromAppError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, post_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, post_errors.ErrInvalidNumPosts) || errors.Is(err, post_errors.ErrInvalidTimestamp):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, post_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, post_errors.ErrUploadFile):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, post_errors.ErrUploadFile):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
