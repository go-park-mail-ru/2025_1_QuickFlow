package file_service

import (
	"context"
	"fmt"
	"io/ioutil"

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
		fileBytes, err = ioutil.ReadAll(file.Reader)
		if err != nil {
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

	resp, err := f.client.UploadFile(ctx, req)
	if err != nil {
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
			fileBytes, err = ioutil.ReadAll(file.Reader)
			if err != nil {
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

	resp, err := f.client.UploadManyFiles(ctx, manyReq)
	if err != nil {
		return nil, fmt.Errorf("fileClient.UploadManyFiles: %w", err)
	}

	return resp.FileUrls, nil
}

// DeleteFile удаляет файл
func (f *FileClient) DeleteFile(ctx context.Context, filename string) error {
	req := &pb.DeleteFileRequest{
		FileUrl: filename,
	}

	_, err := f.client.DeleteFile(ctx, req)
	if err != nil {
		return fmt.Errorf("fileClient.DeleteFile: %w", err)
	}

	return nil
}
