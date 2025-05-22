package dto

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

// ProtoCommentToModel converts proto.Comment to model.Comment
func ProtoCommentToModel(c *pb.Comment) (*models.Comment, error) {
	id, err := uuid.Parse(c.Id)
	if err != nil {
		return nil, err
	}
	userId, err := uuid.Parse(c.UserId)
	if err != nil {
		return nil, err
	}

	postId, err := uuid.Parse(c.PostId)
	if err != nil {
		return nil, err
	}
	return &models.Comment{
		Id:        id,
		PostId:    postId,
		UserId:    userId,
		Text:      c.Text,
		Images:    file_service.ProtoFilesToModels(c.Images),
		CreatedAt: c.CreatedAt.AsTime(),
		UpdatedAt: c.UpdatedAt.AsTime(),
		LikeCount: int(c.LikeCount),
		IsLiked:   c.IsLiked,
	}, nil
}

// ProtoCommentUpdateToModel converts proto.CommentUpdate to model.CommentUpdate
func ProtoCommentUpdateToModel(c *pb.CommentUpdate) (*models.CommentUpdate, error) {
	id, err := uuid.Parse(c.Id)
	if err != nil {
		return nil, err
	}
	return &models.CommentUpdate{
		Id:    id,
		Text:  c.Text,
		Files: file_service.ProtoFilesToModels(c.Files),
	}, nil
}

// ModelCommentToProto converts model.Comment to proto.Comment
func ModelCommentToProto(c *models.Comment) *pb.Comment {
	return &pb.Comment{
		Id:        c.Id.String(),
		PostId:    c.PostId.String(),
		UserId:    c.UserId.String(),
		Text:      c.Text,
		Images:    file_service.ModelFilesToProto(c.Images),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		LikeCount: int64(c.LikeCount),
		IsLiked:   c.IsLiked,
	}
}

func ModelCommentUpdateToProto(c *models.CommentUpdate) *pb.CommentUpdate {
	return &pb.CommentUpdate{
		Id:    c.Id.String(),
		Text:  c.Text,
		Files: file_service.ModelFilesToProto(c.Files),
	}
}
