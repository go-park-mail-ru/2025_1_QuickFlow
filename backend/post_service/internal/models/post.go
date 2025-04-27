package models

import (
	"time"

	"github.com/google/uuid"

	"quickflow/shared/models"
)

type Post struct {
	Id           uuid.UUID
	CreatorId    uuid.UUID
	Desc         string
	Images       []*models.File
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
	Files []*models.File
}
