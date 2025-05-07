package errors

import (
	"errors"
)

var (
	ErrorCommunityNameTooShort       = errors.New("community name is too short")
	ErrorCommunityNameTooLong        = errors.New("community name is too long")
	ErrorCommunityDescriptionTooLong = errors.New("community description is too long")
	ErrorCommunityAvatarSizeExceeded = errors.New("community avatar size exceeded")
	ErrNotFound                      = errors.New("not found")
	ErrNotParticipant                = errors.New("not participant")
	ErrForbidden                     = errors.New("forbidden")
	ErrNilOwnerId                    = errors.New("owner id cannot be nil")
	ErrAlreadyExists                 = errors.New("community with this name already exists")
)
