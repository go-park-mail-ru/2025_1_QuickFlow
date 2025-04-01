package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"quickflow/internal/models"

	"quickflow/internal/delivery/forms"
)

// WriteJSONError sends JSON error response.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(forms.ErrorForm{Error: message})
}

// GetFiles retrieves files from multipart form by key.
func GetFiles(r *http.Request, key string) ([]*models.File, error) {
	var files []*models.File
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

		file.Close()
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
