package usecase

import (
	"context"
	"fmt"

	qf_errors "quickflow/file_service/internal/errors"
	"quickflow/shared/models"
)

type FileStorage interface {
	UploadFile(ctx context.Context, file *models.File) (string, error)
	UploadManyImages(ctx context.Context, files []*models.File) ([]string, error)
	GetFileURL(ctx context.Context, filename string) (string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type FileRepository interface {
	AddFileRecord(ctx context.Context, file *models.File) error
	AddFilesRecords(ctx context.Context, files []*models.File) error
}

type FileValidator interface {
	ValidateFile(file *models.File) error
	ValidateFiles(files []*models.File) error
	ValidateFileName(name string) error
}
type FileUseCase struct {
	fileStorage FileStorage
	fileRepo    FileRepository
	validator   FileValidator
}

// NewFileUseCase creates new file use case.
func NewFileUseCase(fileStorage FileStorage, fileRepo FileRepository, validator FileValidator) *FileUseCase {
	return &FileUseCase{
		fileStorage: fileStorage,
		fileRepo:    fileRepo,
		validator:   validator,
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

	fileUrl, err := f.fileStorage.UploadFile(ctx, fileModel)
	if err != nil {
		return "", fmt.Errorf("f.fileStorage.UploadFile: %w", err)
	}
	fileModel.URL = fileUrl

	// save file record in database
	if err := f.fileRepo.AddFileRecord(ctx, fileModel); err != nil {
		return "", fmt.Errorf("f.fileRepo.AddFileRecord: %w", err)
	}
	return fileUrl, nil
}

func (f *FileUseCase) UploadManyMedia(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileStorage.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileStorage.UploadManyImages: %w", err)
	}

	for i, file := range files {
		file.URL = fileUrls[i]
	}
	// save file records in database
	err = f.fileRepo.AddFilesRecords(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileRepo.AddFilesRecords: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileStorage.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileStorage.UploadManyFiles: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) UploadManyAudios(ctx context.Context, files []*models.File) ([]string, error) {
	// validation
	err := f.validator.ValidateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateFiles: %w", err)
	}

	fileUrls, err := f.fileStorage.UploadManyImages(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("f.fileStorage.UploadManyAudios: %w", err)
	}
	return fileUrls, nil
}

func (f *FileUseCase) GetFileURL(ctx context.Context, filename string) (string, error) {
	// validation
	if err := f.validator.ValidateFileName(filename); err != nil {
		return "", fmt.Errorf("validation.ValidateFileName: %w", err)
	}

	fileUrl, err := f.fileStorage.GetFileURL(ctx, filename)
	if err != nil {
		return "", fmt.Errorf("f.fileStorage.GetFileURL: %w", err)
	}
	return fileUrl, nil
}

func (f *FileUseCase) DeleteFile(ctx context.Context, filename string) error {
	// validation
	if err := f.validator.ValidateFileName(filename); err != nil {
		return fmt.Errorf("validation.ValidateFileName: %w", err)
	}

	err := f.fileStorage.DeleteFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("f.fileStorage.DeleteFile: %w", err)
	}
	return nil
}
