package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"

	"quickflow/post_service/internal/delivery/grpc/dto"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

type CommentUseCase interface {
	FetchCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error)
	AddComment(ctx context.Context, comment models.Comment) (*models.Comment, error)
	DeleteComment(ctx context.Context, userId uuid.UUID, commentId uuid.UUID) error
	UpdateComment(ctx context.Context, update models.CommentUpdate, userId uuid.UUID) (*models.Comment, error)
	LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	GetComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (*models.Comment, error)
	GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error)
}

type CommentServiceServer struct {
	pb.UnimplementedCommentServiceServer
	commentUseCase CommentUseCase
	userUseCase    UserUseCase
}

func NewCommentServiceServer(commentUseCase CommentUseCase, userUseCase UserUseCase) *CommentServiceServer {
	return &CommentServiceServer{
		commentUseCase: commentUseCase,
		userUseCase:    userUseCase,
	}
}

func (c *CommentServiceServer) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentResponse, error) {
	logger.Info(ctx, "AddComment called")
	comment, err := dto.ProtoCommentToModel(req.Comment)
	if err != nil {
		logger.Error(ctx, "Failed to convert proto to model:", err)
		return nil, err
	}

	result, err := c.commentUseCase.AddComment(ctx, *comment)
	if err != nil {
		logger.Error(ctx, "Failed to add comment:", err)
		return nil, err
	}

	return &pb.AddCommentResponse{Comment: dto.ModelCommentToProto(result)}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	logger.Info(ctx, "DeleteComment called")
	commentId, err := uuid.Parse(req.CommentId)
	if err != nil {
		logger.Error(ctx, "Invalid comment ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := c.commentUseCase.DeleteComment(ctx, userId, commentId); err != nil {
		logger.Error(ctx, "Failed to delete comment:", err)
		return nil, err
	}
	return &pb.DeleteCommentResponse{Success: true}, nil
}

func (c *CommentServiceServer) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	logger.Info(ctx, "UpdateComment called")
	update, err := dto.ProtoCommentUpdateToModel(req.Comment)
	if err != nil {
		logger.Error(ctx, "Failed to convert update payload:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	updatedComment, err := c.commentUseCase.UpdateComment(ctx, *update, userId)
	if err != nil {
		logger.Error(ctx, "Failed to update comment:", err)
		return nil, err
	}
	return &pb.UpdateCommentResponse{Comment: dto.ModelCommentToProto(updatedComment)}, nil
}

func (c *CommentServiceServer) FetchCommentsForPost(ctx context.Context, req *pb.FetchCommentsForPostRequest) (*pb.FetchCommentsForPostResponse, error) {
	logger.Info(ctx, "FetchCommentsForPost called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}
	comments, err := c.commentUseCase.FetchCommentsForPost(ctx, postId, int(req.NumComments), req.Timestamp.AsTime())
	if err != nil {
		logger.Error(ctx, "Failed to fetch comments for post:", err)
		return nil, err
	}
	protoComments := make([]*pb.Comment, len(comments))
	for i, comment := range comments {
		protoComments[i] = dto.ModelCommentToProto(&comment)
	}
	return &pb.FetchCommentsForPostResponse{Comments: protoComments}, nil
}

func (c *CommentServiceServer) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentResponse, error) {
	logger.Info(ctx, "LikeComment called")
	commentId, err := uuid.Parse(req.CommentId)
	if err != nil {
		logger.Error(ctx, "Invalid comment ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := c.commentUseCase.LikeComment(ctx, commentId, userId); err != nil {
		logger.Error(ctx, "Failed to like comment:", err)
		return nil, err
	}
	return &pb.LikeCommentResponse{Success: true}, nil
}

func (c *CommentServiceServer) UnlikeComment(ctx context.Context, req *pb.UnlikeCommentRequest) (*pb.UnlikeCommentResponse, error) {
	logger.Info(ctx, "UnlikeComment called")
	commentId, err := uuid.Parse(req.CommentId)
	if err != nil {
		logger.Error(ctx, "Invalid comment ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}
	if err := c.commentUseCase.UnlikeComment(ctx, commentId, userId); err != nil {
		logger.Error(ctx, "Failed to unlike comment:", err)
		return nil, err
	}
	return &pb.UnlikeCommentResponse{Success: true}, nil
}

func (c *CommentServiceServer) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.GetCommentResponse, error) {
	logger.Info(ctx, "GetComment called")
	commentId, err := uuid.Parse(req.CommentId)
	if err != nil {
		logger.Error(ctx, "Invalid comment ID:", err)
		return nil, err
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID:", err)
		return nil, err
	}

	comment, err := c.commentUseCase.GetComment(ctx, commentId, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get comment:", err)
		return nil, err
	}
	return &pb.GetCommentResponse{Comment: dto.ModelCommentToProto(comment)}, nil
}

func (c *CommentServiceServer) GetLastPostComment(ctx context.Context, req *pb.GetLastPostCommentRequest) (*pb.GetLastPostCommentResponse, error) {
	logger.Info(ctx, "GetLastPostComment called")
	postId, err := uuid.Parse(req.PostId)
	if err != nil {
		logger.Error(ctx, "Invalid post ID:", err)
		return nil, err
	}

	comment, err := c.commentUseCase.GetLastPostComment(ctx, postId)
	if err != nil {
		logger.Error(ctx, "Failed to get last comment for post:", err)
		return nil, err
	}
	return &pb.GetLastPostCommentResponse{Comment: dto.ModelCommentToProto(comment)}, nil
}
