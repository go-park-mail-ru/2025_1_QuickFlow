package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"

	"quickflow/gateway/internal/delivery/http/forms"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

var TooManyFilesErr = errors.New("too many files")

// WriteJSONError sends JSON error response.
func WriteJSONError(w http.ResponseWriter, err error) {
	gwErr := errors2.FromGRPCError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(gwErr.HTTPStatus)

	_ = json.NewEncoder(w).Encode(forms.ErrorForm{
		ErrorCode: gwErr.Code,
		Message:   gwErr.Message,
	})
}

func SetRequestId(ctx context.Context) context.Context {
	return context.WithValue(ctx,
		logger.RequestID,
		logger.ReqIdKey(uuid.New().String()))
}

// GetFiles retrieves files from multipart form by key.
func GetFiles(r *http.Request, key string) ([]*models.File, error) {
	var files []*models.File
	// TODO clean code
	if len(r.MultipartForm.File[key]) > 10 {
		return nil, TooManyFilesErr
	}
	for _, fileHeaders := range r.MultipartForm.File[key] {
		mimeType, err := detectMimeType(fileHeaders)
		if err != nil {
			return nil, fmt.Errorf("failed to detect MIME type: %w", err)
		}

		file, err := fileHeaders.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		files = append(files, &models.File{
			Reader:   file,
			Name:     fileHeaders.Filename,
			Size:     fileHeaders.Size,
			Ext:      filepath.Ext(fileHeaders.Filename),
			MimeType: mimeType,
		})

		//file.Close()
	}

	return files, nil
}

func GetFile(r *http.Request, key string) (*models.File, error) {
	fileHeaders := r.MultipartForm.File[key]
	if len(fileHeaders) == 0 {
		return nil, nil
	}

	fileHeader := fileHeaders[0]
	mimeType, err := detectMimeType(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to detect MIME type: %w", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileModel := &models.File{
		Reader:   file,
		Name:     fileHeader.Filename,
		Size:     fileHeader.Size,
		Ext:      filepath.Ext(fileHeader.Filename),
		MimeType: mimeType,
	}

	file.Close()

	return fileModel, nil
}

// detectMimeType определяет MIME-тип файла, сначала проверяя заголовки, затем анализируя содержимое.
func detectMimeType(fileHeader *multipart.FileHeader) (string, error) {
	// Попробуем получить MIME-тип из заголовков
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "" {
		return mimeType, nil
	}

	// Если в заголовках нет, пробуем определить по содержимому
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Читаем первые 512 байтов (это стандартный размер для определения типа)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}
