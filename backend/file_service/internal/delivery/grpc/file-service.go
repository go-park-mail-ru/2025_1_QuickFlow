package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/file_service/internal/delivery/grpc/dto"
	qf_errors "quickflow/file_service/internal/errors"
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
	dtoFile := dto.MapUploadFileRequestToDTO(req)

	fileURL, err := s.fileUC.UploadFile(ctx, dto.MapDTOFileToModel(dtoFile))
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.UploadFileResponse{FileUrl: fileURL}, nil
}

func (s *FileServiceServer) UploadManyFiles(ctx context.Context, req *pb.UploadManyFilesRequest) (*pb.UploadManyFilesResponse, error) {
	dtoFiles := dto.MapUploadManyFilesRequestToDTO(req)

	files := make([]*models.File, len(dtoFiles.Files))
	for i, file := range dtoFiles.Files {
		files[i] = dto.MapDTOFileToModel(file)
	}

	fileURLs, err := s.fileUC.UploadManyFiles(ctx, files)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.UploadManyFilesResponse{FileUrls: fileURLs}, nil
}

func (s *FileServiceServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	err := s.fileUC.DeleteFile(ctx, req.FileUrl)
	if err != nil {
		return &pb.DeleteFileResponse{Success: false}, grpcErrorFromAppError(err)
	}

	return &pb.DeleteFileResponse{Success: true}, nil
}

func grpcErrorFromAppError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, qf_errors.ErrInvalidFileName) ||
		errors.Is(err, qf_errors.ErrInvalidFileSize) ||
		errors.Is(err, qf_errors.ErrUnsupportedFileType) ||
		errors.Is(err, qf_errors.ErrTooManyFiles):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}

}
