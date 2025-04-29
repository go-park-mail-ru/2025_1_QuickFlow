package errors

import (
	"errors"
)

var (
	ErrPostDoesNotBelongToUser = errors.New("post does not belong to user")
	ErrPostNotFound            = errors.New("post not found")
	ErrUploadFile              = errors.New("upload file error")
	ErrInvalidNumPosts         = errors.New("invalid number of posts")
	ErrInvalidTimestamp        = errors.New("invalid timestamp")
	ErrNotFound                = errors.New("not found")
	ErrInvalidUUID             = errors.New("invalid uuid")
)
