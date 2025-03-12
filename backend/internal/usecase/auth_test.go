package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"quickflow/internal/models"
	"testing"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SaveUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) GetUser(ctx context.Context, authData models.LoginData) (models.User, error) {
	args := m.Called(ctx, authData)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUId(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) IsExists(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}

// MockSessionRepository для тестов
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) SaveSession(ctx context.Context, userId uuid.UUID, session models.Session) error {
	args := m.Called(ctx, userId, session)
	return args.Error(0)
}

func (m *MockSessionRepository) LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionRepository) IsExists(ctx context.Context, sessionId uuid.UUID) (bool, error) {
	args := m.Called(ctx, sessionId)
	return args.Bool(0), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, sessionId string) error {
	args := m.Called(ctx, sessionId)
	return args.Error(0)
}

func TestCreateUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	authService := NewAuthService(mockUserRepo, mockSessionRepo)

	ctx := context.Background()
	user := models.User{Login: "testuser", Password: "securepassword"}
	userID := uuid.New()

	mockUserRepo.On("IsExists", ctx, user.Login).Return(false, nil)
	mockUserRepo.On("SaveUser", ctx, mock.Anything).Return(userID, nil)
	mockSessionRepo.On("IsExists", ctx, mock.Anything).Return(false, nil)
	mockSessionRepo.On("SaveSession", ctx, userID, mock.Anything).Return(nil)

	createdUserID, createdSession, err := authService.CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, userID, createdUserID)
	assert.NotEqual(t, uuid.Nil, createdSession.SessionId)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	authService := NewAuthService(mockUserRepo, mockSessionRepo)

	ctx := context.Background()
	user := models.User{Login: "existinguser", Password: "securepassword"}

	mockUserRepo.On("IsExists", ctx, user.Login).Return(true, nil)

	_, _, err := authService.CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	authService := NewAuthService(mockUserRepo, mockSessionRepo)

	ctx := context.Background()
	loginData := models.LoginData{Login: "testuser", Password: "securepassword"}
	user := models.User{Id: uuid.New(), Login: loginData.Login, Password: loginData.Password}

	mockUserRepo.On("GetUser", ctx, loginData).Return(user, nil)
	mockSessionRepo.On("IsExists", ctx, mock.Anything).Return(false, nil)
	mockSessionRepo.On("SaveSession", ctx, user.Id, mock.Anything).Return(nil)

	createdSession, err := authService.GetUser(ctx, loginData)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdSession.SessionId)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestLookupUserSession(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	authService := NewAuthService(mockUserRepo, mockSessionRepo)

	ctx := context.Background()
	session := models.CreateSession()
	userID := uuid.New()
	user := models.User{Id: userID, Login: "testuser"}

	mockSessionRepo.On("LookupUserSession", ctx, session).Return(userID, nil)
	mockUserRepo.On("GetUserByUId", ctx, userID).Return(user, nil)

	foundUser, err := authService.LookupUserSession(ctx, session)

	assert.NoError(t, err)
	assert.Equal(t, user, foundUser)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestDeleteUserSession(t *testing.T) {
	mockSessionRepo := new(MockSessionRepository)
	authService := NewAuthService(nil, mockSessionRepo)

	ctx := context.Background()
	sessionID := uuid.New().String()

	mockSessionRepo.On("DeleteSession", ctx, sessionID).Return(nil)

	err := authService.DeleteUserSession(ctx, sessionID)

	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}
