package models

import (
	"io"
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

type File struct {
	Reader   io.Reader
	Name     string
	Size     int64
	Ext      string
	MimeType string
}

func (f File) String() string {
	return f.Name
}

type PostUpdate struct {
	Id    uuid.UUID
	Desc  string
	Files []*File
}
