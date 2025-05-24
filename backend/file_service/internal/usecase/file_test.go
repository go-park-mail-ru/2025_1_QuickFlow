package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	file_errors "quickflow/file_service/internal/errors"
	"quickflow/file_service/internal/usecase"
	"quickflow/file_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestUploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock the dependencies
	mockRepo := mocks.NewMockFileRepository(ctrl)
	mockValidator := mocks.NewMockFileValidator(ctrl)

	// Define table-driven test cases
	tests := []struct {
		name        string
		file        *models.File
		validator   func()
		repo        func()
		expected    string
		expectedErr error
	}{
		{
			name: "successful upload",
			file: &models.File{
				Name:     "sample.txt",
				Size:     1024,
				Ext:      ".txt",
				MimeType: "text/plain",
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFile(gomock.Any()).Return(nil).Times(1)
			},
			repo: func() {
				mockRepo.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return("http://localhost:9000/sample.txt", nil).Times(1)
			},
			expected:    "http://localhost:9000/sample.txt",
			expectedErr: nil,
		},
		{
			name: "validation error",
			file: &models.File{
				Name:     "sample.txt",
				Size:     1024,
				Ext:      ".txt",
				MimeType: "text/plain",
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFile(gomock.Any()).Return(file_errors.ErrFileIsNil).Times(1)
			},
			repo:        func() {},
			expected:    "",
			expectedErr: file_errors.ErrFileIsNil,
		},
		{
			name: "repo error",
			file: &models.File{
				Name:     "sample.txt",
				Size:     1024,
				Ext:      ".txt",
				MimeType: "text/plain",
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFile(gomock.Any()).Return(nil).Times(1)
			},
			repo: func() {
				mockRepo.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return("", errors.New("upload failed")).Times(1)
			},
			expected:    "",
			expectedErr: errors.New("f.fileStorage.UploadFile: upload failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validator()
			tt.repo()

			// Create FileUseCase
			uc := usecase.NewFileUseCase(mockRepo, mockValidator)

			// Call UploadFile
			result, err := uc.UploadFile(context.Background(), tt.file)

			// Assert the result
			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadManyFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock the dependencies
	mockRepo := mocks.NewMockFileRepository(ctrl)
	mockValidator := mocks.NewMockFileValidator(ctrl)

	// Define table-driven test cases
	tests := []struct {
		name        string
		files       []*models.File
		validator   func()
		repo        func()
		expected    []string
		expectedErr error
	}{
		{
			name: "successful upload many files",
			files: []*models.File{
				{
					Name:     "sample1.txt",
					Size:     1024,
					Ext:      ".txt",
					MimeType: "text/plain",
				},
				{
					Name:     "sample2.txt",
					Size:     1024,
					Ext:      ".txt",
					MimeType: "text/plain",
				},
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFiles(gomock.Any()).Return(nil).Times(1)
			},
			repo: func() {
				mockRepo.EXPECT().UploadManyFiles(gomock.Any(), gomock.Any()).Return([]string{"http://localhost:9000/sample1.txt", "http://localhost:9000/sample2.txt"}, nil).Times(1)
			},
			expected:    []string{"http://localhost:9000/sample1.txt", "http://localhost:9000/sample2.txt"},
			expectedErr: nil,
		},
		{
			name: "validation error",
			files: []*models.File{
				{
					Name:     "sample1.txt",
					Size:     1024,
					Ext:      ".txt",
					MimeType: "text/plain",
				},
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFiles(gomock.Any()).Return(file_errors.ErrFileIsNil).Times(1)
			},
			repo:        func() {},
			expected:    nil,
			expectedErr: file_errors.ErrFileIsNil,
		},
		{
			name: "repo error",
			files: []*models.File{
				{
					Name:     "sample1.txt",
					Size:     1024,
					Ext:      ".txt",
					MimeType: "text/plain",
				},
			},
			validator: func() {
				mockValidator.EXPECT().ValidateFiles(gomock.Any()).Return(nil).Times(1)
			},
			repo: func() {
				mockRepo.EXPECT().UploadManyFiles(gomock.Any(), gomock.Any()).Return(nil, errors.New("upload failed")).Times(1)
			},
			expected:    nil,
			expectedErr: errors.New("f.fileStorage.UploadManyImages: upload failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validator()
			tt.repo()

			// Create FileUseCase
			uc := usecase.NewFileUseCase(mockRepo, mockValidator)

			// Call UploadManyImages
			result, err := uc.UploadManyImages(context.Background(), tt.files)

			// Assert the result
			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
