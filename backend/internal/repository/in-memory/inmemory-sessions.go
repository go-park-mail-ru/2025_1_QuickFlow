package in_memory

import (
	"context"
	"errors"
	"sync"
	"time"

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

// CreateSession creates new session for user.
func (s *InMemorySessionRepository) CreateSession(userId uuid.UUID) models.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionId := uuid.New()

	for _, exists := s.sessions[sessionId]; exists; _, exists = s.sessions[sessionId] {
		sessionId = uuid.New()
	}

	session := models.Session{
		SessionId:  sessionId,
		ExpireDate: time.Now().Add(10 * 24 * time.Hour),
	}

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
