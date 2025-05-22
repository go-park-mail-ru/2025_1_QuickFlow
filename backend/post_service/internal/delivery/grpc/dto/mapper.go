package dto

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	file_service2 "quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

func ProtoPostToModel(p *pb.Post) (*models.Post, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, err
	}
	creatorId, err := uuid.Parse(p.CreatorId)
	if err != nil {
		return nil, err
	}
	return &models.Post{
		Id:           id,
		CreatorId:    creatorId,
		CreatorType:  models.PostCreatorType(p.CreatorType),
		Desc:         p.Description,
		Images:       file_service2.ProtoFilesToModels(p.Images),
		ImagesURL:    p.ImagesUrl,
		CreatedAt:    p.CreatedAt.AsTime(),
		UpdatedAt:    p.UpdatedAt.AsTime(),
		LikeCount:    int(p.LikeCount),
		RepostCount:  int(p.RepostCount),
		CommentCount: int(p.CommentCount),
		IsRepost:     p.IsRepost,
		IsLiked:      p.IsLiked,
	}, nil
}

func ProtoPostUpdateToModel(p *pb.PostUpdate) (*models.PostUpdate, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, err
	}
	return &models.PostUpdate{
		Id:    id,
		Desc:  p.Description,
		Files: file_service2.ProtoFilesToModels(p.Files),
	}, nil
}

func ModelPostToProto(p *models.Post) *pb.Post {
	return &pb.Post{
		Id:           p.Id.String(),
		CreatorId:    p.CreatorId.String(),
		CreatorType:  string(p.CreatorType),
		Description:  p.Desc,
		Images:       file_service2.ModelFilesToProto(p.Images),
		ImagesUrl:    p.ImagesURL,
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
		LikeCount:    int64(p.LikeCount),
		RepostCount:  int64(p.RepostCount),
		CommentCount: int64(p.CommentCount),
		IsRepost:     p.IsRepost,
		IsLiked:      p.IsLiked,
	}
}

func ModelPostUpdateToProto(p *models.PostUpdate) *pb.PostUpdate {
	return &pb.PostUpdate{
		Id:          p.Id.String(),
		Description: p.Desc,
		Files:       file_service2.ModelFilesToProto(p.Files),
	}
}

func ToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
