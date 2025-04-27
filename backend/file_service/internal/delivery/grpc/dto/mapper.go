package dto

import (
	"bytes"
	"path"

	pb "quickflow/file_service/internal/delivery/grpc/proto"
	"quickflow/file_service/internal/models"
)

// mapping to proto
func MapUploadFileRequestToDTO(file *pb.UploadFileRequest) *UploadFileDTO {
	return &UploadFileDTO{
		Name:       file.FileName,
		Size:       file.FileSize,
		Content:    file.File,
		MimeType:   file.FileType,
		AccessMode: AccessModeDTO(file.AccessMode),
		Ext:        path.Ext(file.FileName),
	}
}

func MapDTOToUploadFileResponse(file *UploadFileDTO) *pb.UploadFileResponse {
	return &pb.UploadFileResponse{
		FileUrl: file.URL,
	}
}

func MapUploadManyFilesRequestToDTO(files *pb.UploadManyFilesRequest) *UploadManyFilesDTO {
	var fileDTOs []*UploadFileDTO
	for _, file := range files.Files {
		fileDTOs = append(fileDTOs, MapUploadFileRequestToDTO(file))
	}
	return &UploadManyFilesDTO{
		Files: fileDTOs,
	}
}

func MapDTOToUploadManyFilesResponse(files *UploadManyFilesDTO) *pb.UploadManyFilesResponse {
	var fileURLs []string
	for _, file := range files.Files {
		fileURLs = append(fileURLs, file.URL)
	}
	return &pb.UploadManyFilesResponse{
		FileUrls: fileURLs,
	}
}

func MapDeleteFileRequestToDTO(file *pb.DeleteFileRequest) *DeleteFileDTO {
	return &DeleteFileDTO{
		FileURL: file.FileUrl,
	}
}

func MapDTOToDeleteFileResponse(success bool) *pb.DeleteFileResponse {
	return &pb.DeleteFileResponse{
		Success: success,
	}
}

// mapping to models
func MapDTOFileToModel(fileDTO *UploadFileDTO) *models.File {
	return &models.File{
		Name:       fileDTO.Name,
		Size:       fileDTO.Size,
		Reader:     bytes.NewReader(fileDTO.Content),
		MimeType:   fileDTO.MimeType,
		AccessMode: models.AccessMode(fileDTO.AccessMode),
		Ext:        fileDTO.Ext,
		URL:        fileDTO.URL,
	}
}

func MapModelFileToDTO(file *models.File) *UploadFileDTO {

	return &UploadFileDTO{
		Name:       file.Name,
		Size:       file.Size,
		MimeType:   file.MimeType,
		AccessMode: AccessModeDTO(file.AccessMode),
		Ext:        file.Ext,
		URL:        file.URL,
	}
}
