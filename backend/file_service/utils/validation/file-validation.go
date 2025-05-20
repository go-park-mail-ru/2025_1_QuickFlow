package validation

import (
	"errors"
	"strings"

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

	switch {
	case isVideo(file.MimeType):
		return f.validateFile(file, f.fileConfig.MaxVideoSize, f.fileConfig.AllowedVideoExt)
	case isAudio(file.MimeType):
		return f.validateFile(file, f.fileConfig.MaxAudioSize, f.fileConfig.AllowedAudioExt)
	case isImage(file.MimeType):
		return f.validateFile(file, f.fileConfig.MaxPictureSize, f.fileConfig.AllowedPictureExt)
	default:
		return f.validateFile(file, f.fileConfig.MaxFileSize, f.fileConfig.AllowedFileExt)
	}
}

func (f *FileValidator) validateFile(file *models.File, maxSize int64, allowedExts []string) error {
	if file.Size <= 0 || file.Size > maxSize {
		return qf_errors.ErrInvalidFileSize
	}
	if !isAllowed(strings.ToLower(file.Ext), allowedExts) {
		return qf_errors.ErrUnsupportedFileType
	}
	return nil
}

func isAllowed(ext string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}

func isVideo(mime string) bool {
	return strings.HasPrefix(mime, "video/")
}

func isAudio(mime string) bool {
	return strings.HasPrefix(mime, "audio/")
}

func isImage(mime string) bool {
	return strings.HasPrefix(mime, "image/")
}

func (f *FileValidator) ValidateFiles(files []*models.File) error {
	if len(files) > f.fileConfig.MaxFileCount {
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
