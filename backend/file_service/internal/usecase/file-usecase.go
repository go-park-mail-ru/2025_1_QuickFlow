package usecase

import (
	"context"
	"fmt"

	qf_errors "quickflow/file_service/internal/errors"
	"quickflow/shared/models"
)

type FileRepository interface {
	UploadFile(ctx context.Context, file *models.File) (string, error)
	UploadManyImages(ctx context.Context, files []*models.File) ([]string, error)
	GetFileURL(ctx context.Context, filename string) (string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type FileValidator interface {
	ValidateFile(file *models.File) error
	ValidateFiles(files []*models.File) error
	ValidateFileName(name string) error
}
type FileUseCase struct {
	fileRepo  FileRepository
	validator FileValidator
}

// NewFileUseCase creates new file use case.
func NewFileUseCase(fileRepo FileRepository, validator FileValidator) *FileUseCase {
	return &FileUseCase{
		fileRepo:  fileRepo,
		validator: validator,
	}
}

func (f *FileUseCase) UploadFile(ctx context.Context, fileModel *models.File) (string, error) {
	// validation
	if fileModel == nil {
		return "", qf_errors.ErrFileIsNil
	}
	err := f.validator.ValidateFile(fileModel)
	if err != nil {
		return "", fmt.Errorf("validation.ValidateFile: %w", err)
	}

	fileUrl, err := f.fileRepo.UploadFile(ctx, fileModel)
	if err != nil {
		return "", fmt.Errorf("f.fileRepo.UploadFile: %w", err)
	}
	return fileUrl, nil
}

func (f *FileUseCase) UploadManyMedia(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileRepo.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileRepo.UploadManyImages: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileRepo.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileRepo.UploadManyFiles: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) UploadManyAudios(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileRepo.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileRepo.UploadManyAudios: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) GetFileURL(ctx context.Context, filename string) (string, error) {
	// validation
	if err := f.validator.ValidateFileName(filename); err != nil {
		return "", fmt.Errorf("validation.ValidateFileName: %w", err)
	}

	fileUrl, err := f.fileRepo.GetFileURL(ctx, filename)
	if err != nil {
		return "", fmt.Errorf("f.fileRepo.GetFileURL: %w", err)
	}
	return fileUrl, nil
}

func (f *FileUseCase) DeleteFile(ctx context.Context, filename string) error {
	// validation
	if err := f.validator.ValidateFileName(filename); err != nil {
		return fmt.Errorf("validation.ValidateFileName: %w", err)
	}

	err := f.fileRepo.DeleteFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("f.fileRepo.DeleteFile: %w", err)
	}
	return nil
}
