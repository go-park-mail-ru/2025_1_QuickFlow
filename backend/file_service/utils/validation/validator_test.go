package validation_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "quickflow/file_service/config/validation"
	file_errors "quickflow/file_service/internal/errors"
	"quickflow/file_service/utils/validation"
	"quickflow/shared/models"
)

func TestValidateFile(t *testing.T) {
	// Mock configuration
	mockConfig := &cfg.ValidationConfig{
		MaxFileCount:   10,
		AllowedFileExt: []string{"image/jpeg", "image/png"},
		MaxPictureSize: 5 * 1024 * 1024, // 5MB
	}

	validator := validation.NewFileValidator(mockConfig)

	tests := []struct {
		name        string
		file        *models.File
		expectedErr error
	}{
		{
			name: "invalid file name",
			file: &models.File{
				Name:     "",
				Size:     1024 * 1024, // 1MB
				Ext:      ".jpg",
				MimeType: "image/jpeg",
			},
			expectedErr: file_errors.ErrInvalidFileName,
		},
		{
			name: "invalid file size",
			file: &models.File{
				Name:     "sample.jpg",
				Size:     6 * 1024 * 1024, // 6MB
				Ext:      ".jpg",
				MimeType: "image/jpeg",
			},
			expectedErr: errors.New("file size exceeds the limit"),
		},
		{
			name: "unsupported file type",
			file: &models.File{
				Name:     "sample.txt",
				Size:     1024 * 1024, // 1MB
				Ext:      ".txt",
				MimeType: "text/plain",
			},
			expectedErr: file_errors.ErrUnsupportedFileType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFile(tt.file)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFiles(t *testing.T) {
	// Mock configuration
	mockConfig := &cfg.ValidationConfig{
		MaxFileCount:   10,
		AllowedFileExt: []string{"image/jpeg", "image/png"},
		MaxPictureSize: 5 * 1024 * 1024, // 5MB
	}

	validator := validation.NewFileValidator(mockConfig)

	tests := []struct {
		name        string
		files       []*models.File
		expectedErr error
	}{
		{
			name: "too many files",
			files: []*models.File{
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
				{ // 1st file
					Name:     "sample1.jpg",
					Size:     1024 * 1024, // 1MB
					Ext:      ".jpg",
					MimeType: "image/jpeg",
				},
			},
			expectedErr: file_errors.ErrTooManyFiles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFiles(tt.files)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFileName(t *testing.T) {
	validator := validation.NewFileValidator(&cfg.ValidationConfig{})

	tests := []struct {
		name        string
		filename    string
		expectedErr error
	}{
		{
			name:        "valid file name",
			filename:    "sample.jpg",
			expectedErr: nil,
		},
		{
			name:        "invalid file name",
			filename:    "",
			expectedErr: file_errors.ErrInvalidFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFileName(tt.filename)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
