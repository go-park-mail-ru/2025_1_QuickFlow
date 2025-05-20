package file_service

import (
	"context"
	"fmt"
	"io"
	"sync"

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
	stream, err := f.client.UploadFile(ctx)
	if err != nil {
		return "", fmt.Errorf("UploadFile: %w", err)
	}

	// Сначала метаданные
	err = stream.Send(&pb.UploadFileRequest{
		Data: &pb.UploadFileRequest_Info{
			Info: &pb.File{
				FileName:    file.Name,
				FileType:    file.MimeType,
				FileSize:    file.Size,
				AccessMode:  pb.AccessMode(file.AccessMode),
				DisplayType: string(file.DisplayType),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("send metadata: %w", err)
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := file.Reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}

		err = stream.Send(&pb.UploadFileRequest{
			Data: &pb.UploadFileRequest_Chunk{
				Chunk: buf[:n],
			},
		})
		if err != nil {
			return "", fmt.Errorf("send chunk: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("receive response: %w", err)
	}

	return resp.FileUrl, nil
}

func (f *FileClient) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	stream, err := f.client.UploadManyFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open upload stream: %w", err)
	}

	var (
		mu       sync.Mutex
		fileURLs []string
		wg       sync.WaitGroup
		errChan  = make(chan error, 1)
	)

	// Получаем ответы асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("receive response: %w", err)
				return
			}
			mu.Lock()
			fileURLs = append(fileURLs, resp.FileUrl)
			mu.Unlock()
			logger.Info(ctx, "Uploaded file: %s", resp.FileUrl)
		}
	}()

	// Отправляем файлы
	for _, file := range files {
		err := stream.Send(&pb.UploadFileRequest{
			Data: &pb.UploadFileRequest_Info{
				Info: &pb.File{
					FileName:    file.Name,
					FileType:    file.MimeType,
					FileSize:    file.Size,
					AccessMode:  pb.AccessMode(file.AccessMode),
					DisplayType: string(file.DisplayType),
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("send info: %w", err)
		}

		buf := make([]byte, 32*1024)
		for {
			n, err := file.Reader.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("read file: %s", file.Name))
				return nil, fmt.Errorf("read file: %w", err)
			}
			err = stream.Send(&pb.UploadFileRequest{
				Data: &pb.UploadFileRequest_Chunk{
					Chunk: buf[:n],
				},
			})
			if err != nil {
				return nil, fmt.Errorf("send chunk: %w", err)
			}
		}

		if closer, ok := file.Reader.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				logger.Warn(ctx, "failed to close file %s: %v", file.Name, err)
			}
		}
	}

	// Закрываем отправку (но НЕ соединение!)
	err = stream.CloseSend()
	if err != nil {
		return nil, fmt.Errorf("close send: %w", err)
	}

	// Дожидаемся завершения чтения
	wg.Wait()

	// Проверим ошибки получения
	select {
	case err := <-errChan:
		return nil, err
	default:
		return fileURLs, nil
	}
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
