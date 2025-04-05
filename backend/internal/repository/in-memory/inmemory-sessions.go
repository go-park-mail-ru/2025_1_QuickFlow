package in_memory

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"quickflow/internal/models"
	tsmap "quickflow/pkg/thread-safe-map"
)

type InMemorySessionRepository struct {
	sessions tsmap.ThreadSafeMap[uuid.UUID, uuid.UUID] // sessionId -> userId
}

// NewInMemorySessionRepository creates new storage instance.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: *tsmap.NewThreadSafeMap[uuid.UUID, uuid.UUID](),
	}
}

// SaveSession creates new session for user.
func (s *InMemorySessionRepository) SaveSession(_ context.Context, userId uuid.UUID, session models.Session) models.Session {

	s.sessions.Set(session.SessionId, userId)
	return session
}

// LookupUserSession returns user by session.
func (s *InMemorySessionRepository) LookupUserSession(_ context.Context, session models.Session) (uuid.UUID, error) {

	userId, found := s.sessions.Get(session.SessionId)
	if !found {
		return uuid.Nil, errors.New("session not found")
	}

	return userId, nil
}

func (s *InMemorySessionRepository) IsExists(_ context.Context, sessionId uuid.UUID) (bool, error) {
	if _, ok := s.sessions.Get(sessionId); ok {
		return true, nil
	}

	return false, nil
}
