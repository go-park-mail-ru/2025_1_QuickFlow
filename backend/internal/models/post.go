package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Id        uuid.UUID
	CreatorId uuid.UUID
	Desc      string
	Pics      []string
	CreatedAt time.Time
}
