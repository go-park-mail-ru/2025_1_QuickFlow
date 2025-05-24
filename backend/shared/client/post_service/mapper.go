package post_service

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	file_service2 "quickflow/shared/client/file_service"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

func ProtoPostToModel(p *pb.Post) (*shared_models.Post, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, err
	}
	creatorId, err := uuid.Parse(p.CreatorId)
	if err != nil {
		return nil, err
	}
	return &shared_models.Post{
		Id:           id,
		CreatorId:    creatorId,
		CreatorType:  shared_models.PostCreatorType(p.CreatorType),
		Desc:         p.Description,
		Files:        file_service2.ProtoFilesToModels(p.Files),
		CreatedAt:    p.CreatedAt.AsTime(),
		UpdatedAt:    p.UpdatedAt.AsTime(),
		LikeCount:    int(p.LikeCount),
		RepostCount:  int(p.RepostCount),
		CommentCount: int(p.CommentCount),
		IsRepost:     p.IsRepost,
		IsLiked:      p.IsLiked,
	}, nil
}

func ProtoPostUpdateToModel(p *pb.PostUpdate) (*shared_models.PostUpdate, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, err
	}
	return &shared_models.PostUpdate{
		Id:    id,
		Desc:  p.Description,
		Files: file_service2.ProtoFilesToModels(p.Files),
	}, nil
}

func ModelPostToProto(p *shared_models.Post) *pb.Post {
	return &pb.Post{
		Id:           p.Id.String(),
		CreatorId:    p.CreatorId.String(),
		CreatorType:  string(p.CreatorType),
		Description:  p.Desc,
		Files:        file_service2.ModelFilesToProto(p.Files),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
		LikeCount:    int64(p.LikeCount),
		RepostCount:  int64(p.RepostCount),
		CommentCount: int64(p.CommentCount),
		IsRepost:     p.IsRepost,
		IsLiked:      p.IsLiked,
	}
}

func ModelPostUpdateToProto(p *shared_models.PostUpdate) *pb.PostUpdate {
	return &pb.PostUpdate{
		Id:          p.Id.String(),
		Description: p.Desc,
		Files:       file_service2.ModelFilesToProto(p.Files),
	}
}

func ToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func convertProtoPosts(protoPosts []*pb.Post) ([]shared_models.Post, error) {
	result := make([]shared_models.Post, len(protoPosts))
	for i, p := range protoPosts {
		post, err := ProtoPostToModel(p)
		if err != nil {
			return nil, err
		}
		result[i] = *post
	}
	return result, nil
}

// ProtoCommentToModel converts proto.Comment to model.Comment
func ProtoCommentToModel(c *pb.Comment) (*shared_models.Comment, error) {
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
	return &shared_models.Comment{
		Id:        id,
		PostId:    postId,
		UserId:    userId,
		Text:      c.Text,
		Images:    file_service2.ProtoFilesToModels(c.Images),
		CreatedAt: c.CreatedAt.AsTime(),
		UpdatedAt: c.UpdatedAt.AsTime(),
		LikeCount: int(c.LikeCount),
		IsLiked:   c.IsLiked,
	}, nil
}

// ProtoCommentUpdateToModel converts proto.CommentUpdate to model.CommentUpdate
func ProtoCommentUpdateToModel(c *pb.CommentUpdate) (*shared_models.CommentUpdate, error) {
	id, err := uuid.Parse(c.Id)
	if err != nil {
		return nil, err
	}
	return &shared_models.CommentUpdate{
		Id:    id,
		Text:  c.Text,
		Files: file_service2.ProtoFilesToModels(c.Files),
	}, nil
}

// ModelCommentToProto converts model.Comment to proto.Comment
func ModelCommentToProto(c *shared_models.Comment) *pb.Comment {
	return &pb.Comment{
		Id:        c.Id.String(),
		PostId:    c.PostId.String(),
		UserId:    c.UserId.String(),
		Text:      c.Text,
		Images:    file_service2.ModelFilesToProto(c.Images),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		LikeCount: int64(c.LikeCount),
		IsLiked:   c.IsLiked,
	}
}

func ModelCommentUpdateToProto(c *shared_models.CommentUpdate) *pb.CommentUpdate {
	return &pb.CommentUpdate{
		Id:    c.Id.String(),
		Text:  c.Text,
		Files: file_service2.ModelFilesToProto(c.Files),
	}
}
