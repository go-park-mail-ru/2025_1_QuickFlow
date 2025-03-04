package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionId  uuid.UUID
	ExpireDate time.Time
}

func CreateSession() Session {
	return Session{
		SessionId:  uuid.New(),
		ExpireDate: time.Now().Add(10 * 24 * time.Hour),
	}
}
