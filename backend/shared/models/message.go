package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID
	Text        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attachments []*File

	SenderID   uuid.UUID
	ChatID     uuid.UUID
	ReceiverID uuid.UUID
}
