package post_service

import (
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		Images:       ProtoFilesToModel(p.Images),
		ImagesURL:    p.ImagesUrl,
		CreatedAt:    p.CreatedAt.AsTime(),
		UpdatedAt:    p.UpdatedAt.AsTime(),
		LikeCount:    int(p.LikeCount),
		RepostCount:  int(p.RepostCount),
		CommentCount: int(p.CommentCount),
		IsRepost:     p.IsRepost,
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
		Files: ProtoFilesToModel(p.Files),
	}, nil
}

func ModelFileToProto(f *shared_models.File) *file_service.File {
	if f == nil {
		return nil
	}

	file := &file_service.File{
		FileName:   f.Name,
		FileSize:   f.Size,
		FileType:   f.MimeType,
		AccessMode: file_service.AccessMode(f.AccessMode),
		Url:        f.URL,
	}

	if f.Reader != nil {
		var err error
		file.File, err = io.ReadAll(f.Reader)
		if err != nil {
			return nil
		}

	} else {
		return nil
	}

	return file
}

func ModelFilesToProto(files []*shared_models.File) []*file_service.File {
	result := make([]*file_service.File, len(files))
	for i, f := range files {
		result[i] = ModelFileToProto(f)
	}
	return result
}

func ModelPostToProto(p *shared_models.Post) *pb.Post {
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
	}
}

func ModelPostUpdateToProto(p *shared_models.PostUpdate) *pb.PostUpdate {
	return &pb.PostUpdate{
		Id:          p.Id.String(),
		Description: p.Desc,
		Files:       ModelFilesToProto(p.Files),
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
