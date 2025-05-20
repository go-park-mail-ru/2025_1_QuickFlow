package grpc

import (
	"context"
	"fmt"
	"io"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dto "quickflow/shared/client/file_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
)

type FileUseCase interface {
	UploadFile(ctx context.Context, fileModel *models.File) (string, error)
	UploadManyMedia(ctx context.Context, files []*models.File) ([]string, error)
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

func (s *FileServiceServer) UploadFile(stream pb.FileService_UploadFileServer) error {
	var (
		fileInfo *pb.File
		tempFile *os.File
	)

	ctx := stream.Context()
	logger.Info(ctx, "Started streaming UploadFile request")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			defer os.Remove(tempFile.Name())

			f, err := os.Open(tempFile.Name())
			if err != nil {
				logger.Error(ctx, "Failed to reopen temp file: %v", err)
				return err
			}
			defer f.Close()

			fileURL, err := s.fileUC.UploadFile(ctx, dto.ProtoFileToModel(fileInfo))
			if err != nil {
				logger.Error(ctx, "Upload usecase failed: %v", err)
				return err
			}

			logger.Info(ctx, "File uploaded successfully: %s", fileURL)
			return stream.SendAndClose(&pb.UploadFileResponse{FileUrl: fileURL})
		}

		if err != nil {
			logger.Error(ctx, "Error receiving stream: %v", err)
			return err
		}

		switch x := req.Data.(type) {
		case *pb.UploadFileRequest_Info:
			fileInfo = x.Info
			tempFile, err = os.CreateTemp("", "upload-*")
			if err != nil {
				logger.Error(ctx, "Failed to create temp file: %v", err)
				return err
			}
			defer tempFile.Close()

		case *pb.UploadFileRequest_Chunk:
			if tempFile == nil {
				return status.Errorf(codes.InvalidArgument, "FileInfo must be sent before chunks")
			}
			_, err := tempFile.Write(x.Chunk)
			if err != nil {
				logger.Error(ctx, "Failed to write chunk: %v", err)
				return err
			}
		}
	}
}

func (s *FileServiceServer) UploadManyFiles(stream pb.FileService_UploadManyFilesServer) error {
	var (
		currentInfo *pb.File
		tempFile    *os.File
		ctx         = stream.Context()
	)

	defer func() {
		if tempFile != nil {
			tempFile.Close()
			os.Remove(tempFile.Name()) // чистим tmp
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Обработка последнего файла, если был начат
			if currentInfo != nil && tempFile != nil {
				fileURL, err := s.finalizeUploadedFile(ctx, currentInfo, tempFile)
				if err != nil {
					logger.Error(ctx, "Error finalizing last file: %v", err)
					return err
				}
				if err := stream.Send(&pb.UploadFileResponse{
					FileUrl: fileURL,
				}); err != nil {
					logger.Error(ctx, "Failed to send response: %v", err)
					return err
				}
			}
			logger.Info(ctx, "All files received and processed")
			return nil
		}
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Error receiving upload stream: %v", err))
			return err
		}

		switch data := req.Data.(type) {
		case *pb.UploadFileRequest_Info:
			// Завершаем предыдущий файл
			if currentInfo != nil && tempFile != nil {
				fileURL, err := s.finalizeUploadedFile(ctx, currentInfo, tempFile)
				if err != nil {
					logger.Error(ctx, fmt.Sprintf("Error finalizing file: %v", err))
					return err
				}
				if err := stream.Send(&pb.UploadFileResponse{
					FileUrl: fileURL,
				}); err != nil {
					logger.Error(ctx, fmt.Sprintf("Failed to send response: %v", err))
					return err
				}
				tempFile.Close()
				os.Remove(tempFile.Name())
			}

			// Начинаем новый файл
			currentInfo = data.Info
			tempFile, err = os.CreateTemp("", "upload-*")
			if err != nil {
				logger.Error(ctx, "Failed to create temp file: %v", err)
				return err
			}

		case *pb.UploadFileRequest_Chunk:
			if tempFile == nil {
				return status.Errorf(codes.InvalidArgument, "FileInfo must be sent before chunks")
			}
			_, err := tempFile.Write(data.Chunk)
			if err != nil {
				logger.Error(ctx, "Failed to write chunk: %v", err)
				return err
			}
		}
	}
}

// finalizeUploadedFile обрабатывает файл после его получения
func (s *FileServiceServer) finalizeUploadedFile(
	ctx context.Context,
	info *pb.File,
	tempFile *os.File,
) (string, error) {
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	f, err := os.Open(tempFile.Name())
	if err != nil {
		logger.Error(ctx, "Failed to reopen temp file: %v", err)
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		logger.Error(ctx, "Failed to read file: %v", err)
		return "", err
	}

	info.File = data

	fileURL, err := s.fileUC.UploadFile(ctx, dto.ProtoFileToModel(info))
	if err != nil {
		logger.Error(ctx, "Upload usecase failed: %v", err)
		return "", err
	}

	return fileURL, nil
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
