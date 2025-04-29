package dto

import (
	"bytes"
	"io"
	"path"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "quickflow/post_service/internal/delivery/grpc/proto"
	"quickflow/post_service/internal/models"
	shared_models "quickflow/shared/models"
	file_service "quickflow/shared/proto/file_service"
)

func MapFileToFileDTO(file *shared_models.File) *FileDTO {
	if file == nil {
		return nil
	}
	data := make([]byte, 0)
	if file.Reader != nil {
		data = make([]byte, file.Size)
		_, err := file.Reader.Read(data)
		if err != nil {
			return nil
		}
	}

	return &FileDTO{
		Name:       file.Name,
		Size:       file.Size,
		Ext:        file.Ext,
		MimeType:   file.MimeType,
		AccessMode: AccessModeDTO(file.AccessMode),
		URL:        file.URL,
		Content:    data,
	}
}

func MapFileDTOToFile(fileDTO *FileDTO) *shared_models.File {
	if fileDTO == nil {
		return nil
	}
	var reader io.Reader
	if fileDTO.Content != nil {
		reader = bytes.NewReader(fileDTO.Content)
	}

	return &shared_models.File{
		Name:       fileDTO.Name,
		Size:       fileDTO.Size,
		Ext:        fileDTO.Ext,
		MimeType:   fileDTO.MimeType,
		AccessMode: shared_models.AccessMode(fileDTO.AccessMode),
		URL:        fileDTO.URL,
		Reader:     reader,
	}
}

func (p *PostDTO) ConvertToProto() *pb.Post {
	return &pb.Post{
		Id:           p.Id.String(),
		CreatorId:    p.CreatorId.String(),
		Description:  p.Description,
		ImagesUrl:    p.ImagesURL,
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
		LikeCount:    int64(p.LikeCount),
		RepostCount:  int64(p.RepostCount),
		CommentCount: int64(p.CommentCount),
		IsRepost:     p.IsRepost,
	}
}

func MapProtoFileToDTO(file *file_service.File) *FileDTO {
	if file == nil {
		return nil
	}
	return &FileDTO{
		Name:       file.FileName,
		Size:       file.FileSize,
		Ext:        path.Ext(file.FileName),
		MimeType:   file.FileType,
		AccessMode: AccessModeDTO(file.AccessMode),
	}
}

func ConvertFilesProtoToDTO(files []*file_service.File) []*FileDTO {
	filesDTO := make([]*FileDTO, len(files))
	for i, file := range files {
		filesDTO[i] = MapProtoFileToDTO(file)
	}
	return filesDTO
}

func ConvertFilesDTOToProto(files []*FileDTO) []*file_service.File {
	filesProto := make([]*file_service.File, len(files))
	for i, file := range files {
		filesProto[i] = &file_service.File{
			FileName:   file.Name,
			FileSize:   file.Size,
			FileType:   file.MimeType,
			AccessMode: file_service.AccessMode(file.AccessMode),
		}
	}
	return filesProto
}

func ConvertFromProto(fromProto *pb.Post) (*PostDTO, error) {
	id, err := uuid.Parse(fromProto.Id)
	if err != nil {
		return nil, err
	}
	creatorId, err := uuid.Parse(fromProto.CreatorId)
	if err != nil {
		return nil, err
	}

	return &PostDTO{
		Id:           id,
		CreatorId:    creatorId,
		Description:  fromProto.Description,
		ImagesURL:    fromProto.ImagesUrl,
		Images:       ConvertFilesProtoToDTO(fromProto.Images),
		CreatedAt:    fromProto.CreatedAt.AsTime(),
		UpdatedAt:    fromProto.UpdatedAt.AsTime(),
		LikeCount:    int(fromProto.LikeCount),
		RepostCount:  int(fromProto.RepostCount),
		CommentCount: int(fromProto.CommentCount),
		IsRepost:     fromProto.IsRepost,
	}, nil
}

func (p *PostDTO) ConvertToModel() *models.Post {
	return &models.Post{
		Id:           p.Id,
		CreatorId:    p.CreatorId,
		Desc:         p.Description,
		Images:       ToFilesModel(p.Images),
		ImagesURL:    p.ImagesURL,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		LikeCount:    p.LikeCount,
		RepostCount:  p.RepostCount,
		CommentCount: p.CommentCount,
		IsRepost:     p.IsRepost,
	}
}

func ConvertFromModel(fromProto *models.Post) *PostDTO {
	return &PostDTO{
		Id:           fromProto.Id,
		CreatorId:    fromProto.CreatorId,
		Description:  fromProto.Desc,
		ImagesURL:    fromProto.ImagesURL,
		CreatedAt:    fromProto.CreatedAt,
		UpdatedAt:    fromProto.UpdatedAt,
		LikeCount:    fromProto.LikeCount,
		RepostCount:  fromProto.RepostCount,
		CommentCount: fromProto.CommentCount,
		IsRepost:     fromProto.IsRepost,
	}
}

func UpdateDTOFromProto(fromProto *pb.PostUpdate) (*PostUpdateDTO, error) {
	id, err := uuid.Parse(fromProto.Id)
	if err != nil {
		return nil, err
	}

	return &PostUpdateDTO{
		Id:    id,
		Desc:  fromProto.Description,
		Files: ConvertFilesProtoToDTO(fromProto.Files),
	}, nil
}

func (p *PostUpdateDTO) UpdateProtoFromDTO() *pb.PostUpdate {
	return &pb.PostUpdate{
		Id:          p.Id.String(),
		Description: p.Desc,
		Files:       ConvertFilesDTOToProto(p.Files),
	}
}

func (p *PostUpdateDTO) UpdateToModel() *models.PostUpdate {
	return &models.PostUpdate{
		Id:    p.Id,
		Desc:  p.Desc,
		Files: ToFilesModel(p.Files),
	}
}
