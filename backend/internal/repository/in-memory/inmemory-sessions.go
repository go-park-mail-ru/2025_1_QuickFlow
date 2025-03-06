package in_memory

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]uuid.UUID // sessionId -> userId
}

// NewInMemorySessionRepository creates new storage instance.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[uuid.UUID]uuid.UUID),
	}
}

// SaveSession creates new session for user.
func (s *InMemorySessionRepository) SaveSession(_ context.Context, userId uuid.UUID, session models.Session) models.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionId] = userId

	return session
}

// LookupUserSession returns user by session.
func (s *InMemorySessionRepository) LookupUserSession(_ context.Context, session models.Session) (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userId, found := s.sessions[session.SessionId]
	if !found {
		return uuid.Nil, errors.New("session not found")
	}

	return userId, nil
}

func (s *InMemorySessionRepository) IsExists(_ context.Context, sessionId uuid.UUID) (bool, error) {
	if _, ok := s.sessions[sessionId]; ok {
		return true, nil
	}

	return false, nil
}
