package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatType int

const (
	ChatTypePrivate ChatType = iota
	ChatTypeGroup
)

type ChatCreationInfo struct {
	Name   string
	Type   ChatType
	Avatar *File
}

type Chat struct {
	ID              uuid.UUID
	Name            string
	AvatarURL       string
	Type            ChatType
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastMessage     Message
	LastReadByOther *time.Time
	LastReadByMe    *time.Time
}
