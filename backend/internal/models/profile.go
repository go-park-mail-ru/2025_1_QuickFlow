package models

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserId        uuid.UUID
	Name          string
	Surname       string
	Sex           Sex
	DateOfBirth   time.Time
	Bio           string
	AvatarUrl     string
	BackgroundUrl string
	Avatar        *File
	Background    *File
}
