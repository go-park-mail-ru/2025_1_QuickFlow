package dto

import (
	"bytes"

	shared_models "quickflow/shared/models"
)

type AccessModeDTO int32

const (
	AccessPublic  AccessModeDTO = 0
	AccessPrivate AccessModeDTO = 1
)

type FileDTO struct {
	Content    []byte
	Name       string
	Size       int64
	Ext        string
	MimeType   string
	AccessMode AccessModeDTO
	URL        string
}

func ToFileModel(dto *FileDTO) *shared_models.File {
	return &shared_models.File{
		Reader:     bytes.NewReader(dto.Content),
		Size:       dto.Size,
		Name:       dto.Name,
		Ext:        dto.Ext,
		MimeType:   dto.MimeType,
		URL:        dto.URL,
		AccessMode: shared_models.AccessMode(dto.AccessMode),
	}
}

func ToFilesModel(dtos []*FileDTO) []*shared_models.File {
	files := make([]*shared_models.File, 0, len(dtos))

	for _, dto := range dtos {
		files = append(files, ToFileModel(dto))
	}
	return files
}
