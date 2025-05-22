package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	PostId    uuid.UUID
	Text      string
	Images    []*File
	CreatedAt time.Time
	UpdatedAt time.Time
	LikeCount int
	IsLiked   bool
}

type CommentUpdate struct {
	Id    uuid.UUID
	Text  string
	Files []*File
}
