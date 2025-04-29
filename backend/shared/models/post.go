package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Id           uuid.UUID
	CreatorId    uuid.UUID
	Desc         string
	Images       []*File
	ImagesURL    []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LikeCount    int
	RepostCount  int
	CommentCount int
	IsRepost     bool
}

type PostUpdate struct {
	Id    uuid.UUID
	Desc  string
	Files []*File
}
