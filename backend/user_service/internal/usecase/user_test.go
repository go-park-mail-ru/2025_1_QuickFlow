package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	shared_models "quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/usecase/mocks"
)

func TestUserUseCase_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	userUseCase := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	tests := []struct {
		name            string
		user            shared_models.User
		profile         shared_models.Profile
		mockUserError   error
		mockSaveError   error
		mockSessionErr  error
		expectedErr     bool
		expectedUserId  uuid.UUID
		expectedSession shared_models.Session
	}{
		{
			name: "Error - Username already exists",
			user: shared_models.User{
				Username: "existingUser",
				Password: "password123",
			},
			profile: shared_models.Profile{
				Username: "existingUser",
			},
			mockUserError:   nil,
			mockSaveError:   user_errors.ErrAlreadyExists,
			mockSessionErr:  nil,
			expectedErr:     true,
			expectedUserId:  uuid.Nil,
			expectedSession: shared_models.Session{},
		},
		{
			name: "Error - Profile validation fails",
			user: shared_models.User{
				Username: "newUser",
				Password: "password123",
			},
			profile: shared_models.Profile{
				Username: "",
			},
			mockUserError:   nil,
			mockSaveError:   nil,
			mockSessionErr:  nil,
			expectedErr:     true,
			expectedUserId:  uuid.Nil,
			expectedSession: shared_models.Session{},
		},
		{
			name: "Error - SaveProfile fails",
			user: shared_models.User{
				Username: "newUser",
				Password: "password123",
			},
			profile: shared_models.Profile{
				Username: "newUser",
			},
			mockUserError:   nil,
			mockSaveError:   nil,
			mockSessionErr:  nil,
			expectedErr:     true,
			expectedUserId:  uuid.Nil,
			expectedSession: shared_models.Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			userId, session, err := userUseCase.CreateUser(context.Background(), tt.user, tt.profile)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserId, userId)
				assert.Equal(t, tt.expectedSession.SessionId, session.SessionId)
			}
		})
	}
}

func TestUserUseCase_AuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	userUseCase := NewUserUseCase(mockUserRepo, mockSessionRepo, nil)

	// Используем фиксированные UUID для тестов
	expectedSessionId := uuid.New()

	tests := []struct {
		name            string
		authData        shared_models.LoginData
		mockUserError   error
		mockSessionErr  error
		expectedErr     bool
		expectedSession shared_models.Session
	}{
		{
			name: "Success - Authenticated",
			authData: shared_models.LoginData{
				Username: "user",
				Password: "password123",
			},
			mockUserError:   nil,
			mockSessionErr:  nil,
			expectedErr:     false,
			expectedSession: shared_models.Session{SessionId: expectedSessionId},
		},
		{
			name: "Error - User not found",
			authData: shared_models.LoginData{
				Username: "invalidUser",
				Password: "wrongPassword",
			},
			mockUserError:   user_errors.ErrNotFound,
			mockSessionErr:  nil,
			expectedErr:     true,
			expectedSession: shared_models.Session{},
		},
		{
			name: "Error - Session saving fails",
			authData: shared_models.LoginData{
				Username: "user",
				Password: "password123",
			},
			mockUserError:   nil,
			mockSessionErr:  errors.New("failed to save session"),
			expectedErr:     true,
			expectedSession: shared_models.Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo.EXPECT().GetUser(context.Background(), tt.authData).Return(shared_models.User{}, tt.mockUserError)

			// Мокаем метод SaveSession
			mockSessionRepo.EXPECT().SaveSession(context.Background(), gomock.Any(), gomock.Any()).Return(tt.mockSessionErr)

			session, err := userUseCase.AuthUser(context.Background(), tt.authData)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSession.SessionId, session.SessionId)
			}
		})
	}
}

func TestUserUseCase_GetUserByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userUseCase := NewUserUseCase(mockUserRepo, nil, nil)

	tests := []struct {
		name          string
		username      string
		mockUser      shared_models.User
		mockUserError error
		expectedErr   bool
		expected      shared_models.User
	}{
		{
			name:     "Success - User found",
			username: "user1",
			mockUser: shared_models.User{
				Username: "user1",
				Id:       uuid.New(),
			},
			mockUserError: nil,
			expectedErr:   false,
			expected:      shared_models.User{Username: "user1"},
		},
		{
			name:          "Error - User not found",
			username:      "user2",
			mockUser:      shared_models.User{},
			mockUserError: user_errors.ErrNotFound,
			expectedErr:   true,
			expected:      shared_models.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo.EXPECT().GetUserByUsername(context.Background(), tt.username).Return(tt.mockUser, tt.mockUserError)

			user, err := userUseCase.GetUserByUsername(context.Background(), tt.username)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}
		})
	}
}

func TestUserUseCase_LookupUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	userUseCase := NewUserUseCase(mockUserRepo, mockSessionRepo, nil)

	tests := []struct {
		name             string
		session          shared_models.Session
		mockSessionError error
		mockUserError    error
		expectedErr      bool
		expectedUser     shared_models.User
	}{
		{
			name:             "Success - User session found",
			session:          shared_models.CreateSession(),
			mockSessionError: nil,
			mockUserError:    nil,
			expectedErr:      false,
			expectedUser:     shared_models.User{Username: "user1", Id: uuid.New()},
		},
		{
			name:             "Error - Session lookup failed",
			session:          shared_models.CreateSession(),
			mockSessionError: errors.New("session not found"),
			mockUserError:    nil,
			expectedErr:      true,
			expectedUser:     shared_models.User{},
		},
		{
			name:             "Error - User lookup failed",
			session:          shared_models.CreateSession(),
			mockSessionError: nil,
			mockUserError:    errors.New("user not found"),
			expectedErr:      true,
			expectedUser:     shared_models.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo.EXPECT().LookupUserSession(context.Background(), tt.session).Return(uuid.New(), tt.mockSessionError)
			mockUserRepo.EXPECT().GetUserByUId(context.Background(), gomock.Any()).Return(tt.expectedUser, tt.mockUserError)

			user, err := userUseCase.LookupUserSession(context.Background(), tt.session)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}
