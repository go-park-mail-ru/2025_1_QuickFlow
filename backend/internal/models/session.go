package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionId  uuid.UUID
	ExpireDate time.Time
}

func CreateSession(sessions map[uuid.UUID]User) Session {
	sessionId := uuid.New()

	for _, exists := sessions[sessionId]; exists; _, exists = sessions[sessionId] {
		sessionId = uuid.New()
	}

	return Session{
		SessionId:  sessionId,
		ExpireDate: time.Now().Add(10 * 24 * time.Hour),
	}
}
