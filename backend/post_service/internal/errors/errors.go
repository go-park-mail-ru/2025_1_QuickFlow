package errors

import (
	"errors"
)

var (
	ErrDoesNotBelongToUser = errors.New("does not belong to user")
	ErrPostNotFound        = errors.New("post not found")
	ErrUploadFile          = errors.New("upload file error")
	ErrInvalidNumPosts     = errors.New("invalid number of posts")
	ErrInvalidTimestamp    = errors.New("invalid timestamp")
	ErrNotFound            = errors.New("not found")
	ErrInvalidUUID         = errors.New("invalid uuid")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidNumComments  = errors.New("invalid number of comments")
)
