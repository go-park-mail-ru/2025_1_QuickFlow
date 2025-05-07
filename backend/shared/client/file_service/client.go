package file_service

import (
	"context"
	"fmt"
	"io"

	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"

	"google.golang.org/grpc"
)

type FileClient struct {
	client pb.FileServiceClient
}

// NewFileClient создаёт новый gRPC клиент
func NewFileClient(conn *grpc.ClientConn) *FileClient {
	return &FileClient{
		client: pb.NewFileServiceClient(conn),
	}
}

// UploadFile загружает один файл
func (f *FileClient) UploadFile(ctx context.Context, file *models.File) (string, error) {
	var fileBytes []byte
	var err error

	// Если есть Reader — читаем
	if file.Reader != nil {
		fileBytes, err = io.ReadAll(file.Reader)
		if err != nil {
			logger.Error(ctx, "Failed to read from file reader: %v", err)
			return "", fmt.Errorf("failed to read file: %w", err)
		}
	}

	req := &pb.UploadFileRequest{
		File: &pb.File{
			FileName:   file.Name,
			File:       fileBytes,
			FileType:   file.MimeType,
			FileSize:   file.Size,
			AccessMode: pb.AccessMode(file.AccessMode), // Явное преобразование enum
		},
	}

	logger.Info(ctx, "Sending request to file_service to upload file: %s", file.Name)

	resp, err := f.client.UploadFile(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to upload file to file_service: %s", file.Name)
		return "", fmt.Errorf("fileClient.UploadFile: %w", err)
	}

	return resp.FileUrl, nil
}

// UploadManyFiles загружает несколько файлов
func (f *FileClient) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	var requests []*pb.UploadFileRequest

	for _, file := range files {
		var fileBytes []byte
		var err error

		if file.Reader != nil {
			fileBytes, err = io.ReadAll(file.Reader)
			if err != nil {
				logger.Error(ctx, "Failed to read from file reader: %v", err)
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
		}

		req := &pb.UploadFileRequest{
			File: &pb.File{
				FileName:   file.Name,
				File:       fileBytes,
				FileType:   file.MimeType,
				FileSize:   file.Size,
				AccessMode: pb.AccessMode(file.AccessMode),
			},
		}

		requests = append(requests, req)
	}

	manyReq := &pb.UploadManyFilesRequest{
		Files: requests,
	}

	logger.Info(ctx, "Sending request to file_service to upload multiple (%d) files", len(requests))
	resp, err := f.client.UploadManyFiles(ctx, manyReq)
	if err != nil {
		logger.Error(ctx, "Failed to upload multiple files to file_service")
		return nil, fmt.Errorf("fileClient.UploadManyFiles: %w", err)
	}

	return resp.FileUrls, nil
}

// DeleteFile удаляет файл
func (f *FileClient) DeleteFile(ctx context.Context, filename string) error {
	req := &pb.DeleteFileRequest{
		FileUrl: filename,
	}

	logger.Info(ctx, "Sending request to file_service to delete file: %s", filename)
	_, err := f.client.DeleteFile(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to delete file from file_service: %s", filename)
		return fmt.Errorf("fileClient.DeleteFile: %w", err)
	}

	return nil
}
