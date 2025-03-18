package in_memory

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

func TestInMemorySessionRepository_SaveSession(t *testing.T) {
	repo := NewInMemorySessionRepository()

	tests := []struct {
		name    string
		userId  uuid.UUID
		session models.Session
		wantErr bool
	}{
		{
			name:    "Save valid session",
			userId:  uuid.New(),
			session: models.Session{SessionId: uuid.New()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = repo.SaveSession(context.Background(), tt.userId, tt.session)
		})
	}
}

func TestInMemorySessionRepository_LookupUserSession(t *testing.T) {
	repo := NewInMemorySessionRepository()

	userId := uuid.New()
	session := models.Session{SessionId: uuid.New()}
	repo.SaveSession(context.Background(), userId, session)

	tests := []struct {
		name       string
		session    models.Session
		expectedId uuid.UUID
		wantErr    bool
	}{
		{
			name:       "Valid session lookup",
			session:    session,
			expectedId: userId,
			wantErr:    false,
		},
		{
			name:       "Invalid session lookup",
			session:    models.Session{SessionId: uuid.New()},
			expectedId: uuid.Nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.LookupUserSession(context.Background(), tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupUserSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.expectedId {
				t.Errorf("LookupUserSession() got = %v, want %v", got, tt.expectedId)
			}
		})
	}
}

func TestInMemorySessionRepository_IsExists(t *testing.T) {
	repo := NewInMemorySessionRepository()

	sessionId := uuid.New()
	// Сначала сохраняем сессию
	session := models.Session{SessionId: sessionId}
	repo.SaveSession(context.Background(), uuid.New(), session)

	tests := []struct {
		name      string
		sessionId uuid.UUID
		wantExist bool
		wantErr   bool
	}{
		{
			name:      "Session exists",
			sessionId: sessionId,
			wantExist: true,
			wantErr:   false,
		},
		{
			name:      "Session does not exist",
			sessionId: uuid.New(),
			wantExist: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.IsExists(context.Background(), tt.sessionId)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if exists != tt.wantExist {
				t.Errorf("IsExists() got = %v, want %v", exists, tt.wantExist)
			}
		})
	}
}
