package validation

import (
	"errors"

	cfg "quickflow/file_service/config/validation"
	qf_errors "quickflow/file_service/internal/errors"
	"quickflow/shared/models"
)

type FileValidator struct {
	fileConfig *cfg.ValidationConfig
}

func NewFileValidator(fileConfig *cfg.ValidationConfig) *FileValidator {
	return &FileValidator{
		fileConfig: fileConfig,
	}
}

func (f *FileValidator) ValidateFile(file *models.File) error {
	if file == nil {
		return errors.New("file is nil")
	}

	if err := f.ValidateFileName(file.Name); err != nil {
		return err
	}
	if err := f.validateFileSize(file.Size); err != nil {
		return err
	}
	if err := f.validateFileMimeType(file.MimeType); err != nil {
		return err
	}
	return nil
}

func (f *FileValidator) ValidateFiles(files []*models.File) error {
	if len(files) == 0 {
		return errors.New("file is empty")
	}

	if len(files) > f.fileConfig.MaxPicturesCount {
		return qf_errors.ErrTooManyFiles
	}

	for _, file := range files {
		if err := f.ValidateFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileValidator) ValidateFileName(name string) error {
	if len(name) == 0 {
		return qf_errors.ErrInvalidFileName
	}
	return nil
}

func (f *FileValidator) validateFileSize(size int64) error {
	if size <= 0 {
		return qf_errors.ErrInvalidFileSize
	}

	if size > f.fileConfig.MaxPictureSize {
		return errors.New("file size exceeds the limit")
	}
	return nil
}

func (f *FileValidator) validateFileMimeType(mimeType string) error {
	if len(mimeType) == 0 {
		return qf_errors.ErrUnsupportedFileType
	}

	for _, ext := range f.fileConfig.AllowedFileExt {
		if mimeType == ext {
			return nil
		}
	}
	return qf_errors.ErrUnsupportedFileType
}
