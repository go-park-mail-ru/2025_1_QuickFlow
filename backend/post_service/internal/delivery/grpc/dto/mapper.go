package dto

import (
	"bytes"
	"io"

	"quickflow/shared/models"
)

func MapFileToFileDTO(file *models.File) *FileDTO {
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

func MapFileDTOToFile(fileDTO *FileDTO) *models.File {
	if fileDTO == nil {
		return nil
	}
	var reader io.Reader
	if fileDTO.Content != nil {
		reader = bytes.NewReader(fileDTO.Content)
	}

	return &models.File{
		Name:       fileDTO.Name,
		Size:       fileDTO.Size,
		Ext:        fileDTO.Ext,
		MimeType:   fileDTO.MimeType,
		AccessMode: models.AccessMode(fileDTO.AccessMode),
		URL:        fileDTO.URL,
		Reader:     reader,
	}
}
