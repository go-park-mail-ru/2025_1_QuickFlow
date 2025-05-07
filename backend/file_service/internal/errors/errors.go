package errors

import (
	"errors"
)

var (
	ErrInvalidFileName     = errors.New("invalid file name")
	ErrInvalidFileSize     = errors.New("invalid file size")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrTooManyFiles        = errors.New("too many files")
	ErrFileIsNil           = errors.New("file is nil")
)
