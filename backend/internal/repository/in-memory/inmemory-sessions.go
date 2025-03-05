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
func (s *InMemorySessionRepository) SaveSession(userId uuid.UUID, session models.Session) models.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionId] = userId

	return session
}

// LookupUserSession returns user by session.
func (s *InMemorySessionRepository) LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userId, found := s.sessions[session.SessionId]
	if !found {
		return uuid.Nil, errors.New("session not found")
	}

	return userId, nil
}

func (s *InMemorySessionRepository) IsExists(sessionId uuid.UUID) bool {
	if _, ok := s.sessions[sessionId]; ok {
		return true
	}

	return false
}
