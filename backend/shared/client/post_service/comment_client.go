package post_service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/timestamppb"

    "quickflow/shared/logger"
    "quickflow/shared/models"
    pb "quickflow/shared/proto/post_service"
)

type CommentClient struct {
    client pb.CommentServiceClient
}

// NewCommentClient создает новый gRPC клиент для комментариев
func NewCommentClient(conn *grpc.ClientConn) *CommentClient {
    return &CommentClient{
        client: pb.NewCommentServiceClient(conn),
    }
}

// FetchCommentsForPost получает комментарии для поста.
func (c *CommentClient) FetchCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error) {
    req := &pb.FetchCommentsForPostRequest{
        PostId:      postId.String(),
        NumComments: int32(numComments),
        Timestamp:   timestamppb.New(timestamp),
    }

    resp, err := c.client.FetchCommentsForPost(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to fetch comments for post:", err)
        return nil, fmt.Errorf("failed to fetch comments for post: %w", err)
    }

    var comments []models.Comment
    for _, protoComment := range resp.Comments {
        comment, err := ProtoCommentToModel(protoComment)
        if err != nil {
            logger.Error(ctx, "Failed to convert proto comment to model:", err)
            continue
        }
        comments = append(comments, *comment)
    }

    return comments, nil
}

// AddComment добавляет новый комментарий.
func (c *CommentClient) AddComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
    protoComment := ModelCommentToProto(&comment)

    req := &pb.AddCommentRequest{
        Comment: protoComment,
    }

    resp, err := c.client.AddComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to add comment:", err)
        return nil, fmt.Errorf("failed to add comment: %w", err)
    }

    return ProtoCommentToModel(resp.Comment)
}

// DeleteComment удаляет комментарий.
func (c *CommentClient) DeleteComment(ctx context.Context, userId uuid.UUID, commentId uuid.UUID) error {
    req := &pb.DeleteCommentRequest{
        CommentId: commentId.String(),
        UserId:    userId.String(),
    }

    _, err := c.client.DeleteComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to delete comment:", err)
        return fmt.Errorf("failed to delete comment: %w", err)
    }

    return nil
}

// UpdateComment обновляет комментарий.
func (c *CommentClient) UpdateComment(ctx context.Context, commentUpdate models.CommentUpdate, userId uuid.UUID) (*models.Comment, error) {
    protoUpdate := ModelCommentUpdateToProto(&commentUpdate)

    req := &pb.UpdateCommentRequest{
        Comment: protoUpdate,
        UserId:  userId.String(),
    }

    resp, err := c.client.UpdateComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to update comment:", err)
        return nil, fmt.Errorf("failed to update comment: %w", err)
    }

    return ProtoCommentToModel(resp.Comment)
}

// LikeComment ставит лайк на комментарий.
func (c *CommentClient) LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
    req := &pb.LikeCommentRequest{
        CommentId: commentId.String(),
        UserId:    userId.String(),
    }

    _, err := c.client.LikeComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to like comment:", err)
        return fmt.Errorf("failed to like comment: %w", err)
    }

    return nil
}

// UnlikeComment убирает лайк с комментария.
func (c *CommentClient) UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
    req := &pb.UnlikeCommentRequest{
        CommentId: commentId.String(),
        UserId:    userId.String(),
    }

    _, err := c.client.UnlikeComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to unlike comment:", err)
        return fmt.Errorf("failed to unlike comment: %w", err)
    }

    return nil
}

// GetComment получает один комментарий.
func (c *CommentClient) GetComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (*models.Comment, error) {
    req := &pb.GetCommentRequest{
        CommentId: commentId.String(),
        UserId:    userId.String(),
    }

    resp, err := c.client.GetComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to get comment:", err)
        return nil, fmt.Errorf("failed to get comment: %w", err)
    }

    return ProtoCommentToModel(resp.Comment)
}

func (c *CommentClient) GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error) {
    req := &pb.GetLastPostCommentRequest{
        PostId: postId.String(),
    }

    resp, err := c.client.GetLastPostComment(ctx, req)
    if err != nil {
        logger.Error(ctx, "Failed to get last post comment:", err)
        return nil, err
    }

    return ProtoCommentToModel(resp.Comment)
}
