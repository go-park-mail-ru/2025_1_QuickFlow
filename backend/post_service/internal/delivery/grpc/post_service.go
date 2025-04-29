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
	pb "quickflow/post_service/internal/delivery/grpc/proto"
	post_errors "quickflow/post_service/internal/errors"
	"quickflow/post_service/internal/models"
	shared_models "quickflow/shared/models"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, user shared_models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchRecommendations(ctx context.Context, user shared_models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	FetchUserPosts(ctx context.Context, user shared_models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (models.Post, error)
	DeletePost(ctx context.Context, user shared_models.User, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (models.Post, error)
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
	dtoPost, err := dto.ConvertFromProto(req.Post)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	post, err := p.postUseCase.AddPost(ctx, *dtoPost.ConvertToModel())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.AddPostResponse{Post: dto.ConvertFromModel(&post).ConvertToProto()}, nil
}

func (p *PostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid post id: %w", post_errors.ErrInvalidUUID, err))
	}

	user, ok := ctx.Value("user").(shared_models.User)
	if !ok {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	err = p.postUseCase.DeletePost(ctx, user, postId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.DeletePostResponse{
		Success: true,
	}, nil
}

func (p *PostServiceServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	post, err := dto.UpdateDTOFromProto(req.Post)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	user, ok := ctx.Value("user").(shared_models.User)
	if !ok {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	updatedPost, err := p.postUseCase.UpdatePost(ctx, *post.UpdateToModel(), user.Id)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.UpdatePostResponse{
		Post: dto.ConvertFromModel(&updatedPost).ConvertToProto(),
	}, nil
}

func (p *PostServiceServer) FetchFeed(ctx context.Context, req *pb.FetchFeedRequest) (*pb.FetchFeedResponse, error) {
	user, ok := ctx.Value("user").(shared_models.User)
	if !ok {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	posts, err := p.postUseCase.FetchFeed(ctx, user, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	dtoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		dtoPosts[i] = dto.ConvertFromModel(&post).ConvertToProto()
	}

	return &pb.FetchFeedResponse{Posts: dtoPosts}, nil
}

func (p *PostServiceServer) FetchRecommendations(ctx context.Context, req *pb.FetchRecommendationsRequest) (*pb.FetchRecommendationsResponse, error) {
	user, ok := ctx.Value("user").(shared_models.User)
	if !ok {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user", post_errors.ErrInvalidUUID))
	}

	posts, err := p.postUseCase.FetchRecommendations(ctx, user, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	dtoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		dtoPosts[i] = dto.ConvertFromModel(&post).ConvertToProto()
	}

	return &pb.FetchRecommendationsResponse{Posts: dtoPosts}, nil
}

func (p *PostServiceServer) FetchUserPosts(ctx context.Context, req *pb.FetchUserPostsRequest) (*pb.FetchUserPostsResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("%w: invalid user id: %w", post_errors.ErrInvalidUUID, err))
	}
	// get user by userId
	user, err := p.userUseCase.GetUserById(ctx, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	posts, err := p.postUseCase.FetchUserPosts(ctx, *user, int(req.NumPosts), req.Timestamp.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	dtoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		dtoPosts[i] = dto.ConvertFromModel(&post).ConvertToProto()
	}

	return &pb.FetchUserPostsResponse{Posts: dtoPosts}, nil
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
