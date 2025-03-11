package http

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"quickflow/internal/models"
	"time"
)

// MockPostUseCase - мок для PostUseCase
type MockPostUseCase struct {
	mock.Mock
}

func (m *MockPostUseCase) AddPost(ctx context.Context, post models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostUseCase) FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error) {
	args := m.Called(ctx, user, numPosts, timestamp)
	return args.Get(0).([]models.Post), args.Error(1)
}

type MockAuthUseCase struct {
	mock.Mock
}

func (m MockAuthUseCase) CreateUser(ctx context.Context, user models.User) (uuid.UUID, models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthUseCase) GetUser(ctx context.Context, authData models.LoginData) (models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthUseCase) LookupUserSession(ctx context.Context, session models.Session) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthUseCase) DeleteUserSession(ctx context.Context, session string) error {
	//TODO implement me
	panic("implement me")
}
