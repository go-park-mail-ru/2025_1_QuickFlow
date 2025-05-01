package file_service

import (
	"bytes"

	shared_models "quickflow/shared/models"
	"quickflow/shared/proto/file_service"
)

func ProtoFileToModel(file *file_service.File) *shared_models.File {
	return &shared_models.File{
		Name:       file.FileName,
		Size:       file.FileSize,
		MimeType:   file.FileType,
		AccessMode: shared_models.AccessMode(file.AccessMode),
		URL:        file.Url,
		Reader:     bytes.NewReader(file.File),
	}
}

func ModelFileToProto(file *shared_models.File) *file_service.File {
	return &file_service.File{
		FileName:   file.Name,
		FileSize:   file.Size,
		FileType:   file.MimeType,
		AccessMode: file_service.AccessMode(file.AccessMode),
		Url:        file.URL,
	}
}

func ProtoFilesToModels(files []*file_service.File) []*shared_models.File {
	res := make([]*shared_models.File, len(files))
	for i, file := range files {
		res[i] = ProtoFileToModel(file)
	}
	return res
}

func FileModelToProto(file *shared_models.File) *file_service.File {
	return &file_service.File{
		FileName:   file.Name,
		FileSize:   file.Size,
		FileType:   file.MimeType,
		AccessMode: file_service.AccessMode(file.AccessMode),
	}
}
