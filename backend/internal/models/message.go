package models

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        uuid.UUID
	Text      string
	CreatedAt time.Time
	IsRead    bool

	SenderID uuid.UUID
	ChatID   uuid.UUID
}
