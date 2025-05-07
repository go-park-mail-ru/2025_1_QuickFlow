package grpc

import (
	"context"
	"quickflow/file_service/internal/delivery/grpc/dto"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
)

type FileUseCase interface {
	UploadFile(ctx context.Context, fileModel *models.File) (string, error)
	UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error)
	GetFileURL(ctx context.Context, filename string) (string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	fileUC FileUseCase
}

func NewFileServiceServer(fileUC FileUseCase) *FileServiceServer {
	return &FileServiceServer{fileUC: fileUC}
}

func (s *FileServiceServer) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	logger.Info(ctx, "Received UploadFile request")

	dtoFile := dto.MapUploadFileRequestToDTO(req)
	fileURL, err := s.fileUC.UploadFile(ctx, dto.MapDTOFileToModel(dtoFile))
	if err != nil {
		logger.Error(ctx, "Failed to upload file: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully uploaded file")
	return &pb.UploadFileResponse{FileUrl: fileURL}, nil
}

func (s *FileServiceServer) UploadManyFiles(ctx context.Context, req *pb.UploadManyFilesRequest) (*pb.UploadManyFilesResponse, error) {
	logger.Info(ctx, "Received UploadManyFiles request")

	dtoFiles := dto.MapUploadManyFilesRequestToDTO(req)
	files := make([]*models.File, len(dtoFiles.Files))
	for i, file := range dtoFiles.Files {
		files[i] = dto.MapDTOFileToModel(file)
	}

	fileURLs, err := s.fileUC.UploadManyFiles(ctx, files)
	if err != nil {
		logger.Error(ctx, "Failed to upload many files: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully uploaded multiple files")
	return &pb.UploadManyFilesResponse{FileUrls: fileURLs}, nil
}

func (s *FileServiceServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	logger.Info(ctx, "Received DeleteFile request")

	err := s.fileUC.DeleteFile(ctx, req.FileUrl)
	if err != nil {
		logger.Error(ctx, "Failed to delete file: ", err)
		return &pb.DeleteFileResponse{Success: false}, err
	}

	logger.Info(ctx, "Successfully deleted file")
	return &pb.DeleteFileResponse{Success: true}, nil
}
