package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID
	Text           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsRead         bool
	Attachments    []*File
	AttachmentURLs []string

	SenderID   uuid.UUID
	ChatID     uuid.UUID
	ReceiverID uuid.UUID
}
