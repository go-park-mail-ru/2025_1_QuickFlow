package dto

import (
	"bytes"
	"path"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	shared_models "quickflow/shared/models"
	"quickflow/shared/proto/file_service"
	pb "quickflow/shared/proto/post_service"
)

func ProtoFileToModel(f *file_service.File) *shared_models.File {
	if f == nil {
		return nil
	}
	return &shared_models.File{
		Name:       f.FileName,
		Size:       f.FileSize,
		Ext:        path.Ext(f.FileName),
		MimeType:   f.FileType,
		AccessMode: shared_models.AccessMode(f.AccessMode),
		Reader:     bytes.NewReader(f.File),
		URL:        f.Url,
	}
}

func ProtoFilesToModel(files []*file_service.File) []*shared_models.File {
	result := make([]*shared_models.File, len(files))
	for i, f := range files {
		result[i] = ProtoFileToModel(f)
	}
	return result
}

// Convert from proto.Post to model.Post
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
		Images:       ProtoFilesToModel(p.Images),
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
		Files: ProtoFilesToModel(p.Files),
	}, nil
}

func ModelFileToProto(f *shared_models.File) *file_service.File {
	if f == nil {
		return nil
	}
	return &file_service.File{
		FileName:   f.Name,
		FileSize:   f.Size,
		FileType:   f.MimeType,
		AccessMode: file_service.AccessMode(f.AccessMode),
		Url:        f.URL,
	}
}

func ModelFilesToProto(files []*shared_models.File) []*file_service.File {
	result := make([]*file_service.File, len(files))
	for i, f := range files {
		result[i] = ModelFileToProto(f)
	}
	return result
}

func ModelPostToProto(p *models.Post) *pb.Post {
	return &pb.Post{
		Id:           p.Id.String(),
		CreatorId:    p.CreatorId.String(),
		CreatorType:  string(p.CreatorType),
		Description:  p.Desc,
		Images:       ModelFilesToProto(p.Images),
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
		Files:       ModelFilesToProto(p.Files),
	}
}

func ToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
