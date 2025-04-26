package usecase_test

import (
	"context"
	"errors"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/internal/usecase"
	"quickflow/monolith/internal/usecase/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	profileRepo := mocks.NewMockProfileRepository(ctrl)

	authService := usecase.NewAuthService(userRepo, sessionRepo, profileRepo)

	ctx := context.Background()
	testUser := models.User{Username: "testuser", Password: "password"}
	testProfile := models.Profile{}
	testUserId := uuid.New()

	tests := []struct {
		name         string
		mockSetup   func()
		user        models.User
		profile     models.Profile
		expectedErr error
		expectedUser bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				userRepo.EXPECT().IsExists(ctx, testUser.Username).Return(false, nil)
				userRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(testUserId, nil)
				profileRepo.EXPECT().SaveProfile(ctx, gomock.Any()).Return(nil)
				sessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil)
				sessionRepo.EXPECT().SaveSession(ctx, testUserId, gomock.Any()).Return(nil)
			},
			user:         testUser,
			profile:      testProfile,
			expectedErr:  nil,
			expectedUser: true,
		},
		{
			name: "User already exists",
			mockSetup: func() {
				userRepo.EXPECT().IsExists(ctx, testUser.Username).Return(true, nil)
			},
			user:         testUser,
			profile:      testProfile,
			expectedErr:  usecase.ErrAlreadyExists,
			expectedUser: false,
		},
		{
			name: "Error checking user existence",
			mockSetup: func() {
				userRepo.EXPECT().IsExists(ctx, testUser.Username).Return(false, errors.New("db error"))
			},
			user:         testUser,
			profile:      testProfile,
			expectedErr:  errors.New("a.userRepo.IsExists: db error"),
			expectedUser: false,
		},
		{
			name: "Error saving user",
			mockSetup: func() {
				userRepo.EXPECT().IsExists(ctx, testUser.Username).Return(false, nil)
				userRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(uuid.Nil, errors.New("save error"))
			},
			user:         testUser,
			profile:      testProfile,
			expectedErr:  errors.New("a.userRepo.SaveUser: save error"),
			expectedUser: false,
		},
		{
			name: "Error saving profile",
			mockSetup: func() {
				userRepo.EXPECT().IsExists(ctx, testUser.Username).Return(false, nil)
				userRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(testUserId, nil)
				profileRepo.EXPECT().SaveProfile(ctx, gomock.Any()).Return(errors.New("profile error"))
			},
			user:         testUser,
			profile:      testProfile,
			expectedErr:  errors.New("a.profileRepo.SaveProfile: profile error"),
			expectedUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			userId, session, err := authService.CreateUser(ctx, tt.user, tt.profile)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedUser {
				assert.NotEqual(t, uuid.Nil, userId)
				assert.NotEqual(t, uuid.Nil, session.SessionId)
			} else {
				assert.Equal(t, uuid.Nil, userId)
			}
		})
	}
}

func TestAuthService_AuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	profileRepo := mocks.NewMockProfileRepository(ctrl)

	authService := usecase.NewAuthService(userRepo, sessionRepo, profileRepo)

	ctx := context.Background()
	testUser := models.User{Id: uuid.New(), Username: "testuser", Password: "password"}
	testAuthData := models.LoginData{Login: "testuser", Password: "password"}

	tests := []struct {
		name         string
		mockSetup   func()
		authData    models.LoginData
		expectedErr error
		expectedSess bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				userRepo.EXPECT().GetUser(ctx, testAuthData).Return(testUser, nil)
				sessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil)
				sessionRepo.EXPECT().SaveSession(ctx, testUser.Id, gomock.Any()).Return(nil)
			},
			authData:     testAuthData,
			expectedErr:  nil,
			expectedSess: true,
		},
		{
			name: "User not found",
			mockSetup: func() {
				userRepo.EXPECT().GetUser(ctx, testAuthData).Return(models.User{}, usecase.ErrNotFound)
			},
			authData:     testAuthData,
			expectedErr:  errors.New("a.userRepo.GetUser: not found"),
			expectedSess: false,
		},
		{
			name: "Error checking session existence",
			mockSetup: func() {
				userRepo.EXPECT().GetUser(ctx, testAuthData).Return(testUser, nil)
				sessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, errors.New("session error"))
			},
			authData:     testAuthData,
			expectedErr:  errors.New("a.sessionRepo.IsExists: session error"),
			expectedSess: false,
		},
		{
			name: "Error saving session",
			mockSetup: func() {
				userRepo.EXPECT().GetUser(ctx, testAuthData).Return(testUser, nil)
				sessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil)
				sessionRepo.EXPECT().SaveSession(ctx, testUser.Id, gomock.Any()).Return(errors.New("save error"))
			},
			authData:     testAuthData,
			expectedErr:  errors.New("a.sessionRepo.SaveSession: save error"),
			expectedSess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			session, err := authService.AuthUser(ctx, tt.authData)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedSess {
				assert.NotEqual(t, uuid.Nil, session.SessionId)
			} else {
				assert.Equal(t, models.Session{}, session)
			}
		})
	}
}

func TestAuthService_LookupUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	profileRepo := mocks.NewMockProfileRepository(ctrl)

	authService := usecase.NewAuthService(userRepo, sessionRepo, profileRepo)

	ctx := context.Background()
	testUser := models.User{Id: uuid.New(), Username: "testuser"}
	testSession := models.Session{SessionId: uuid.New()}

	tests := []struct {
		name        string
		mockSetup   func()
		session     models.Session
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				sessionRepo.EXPECT().LookupUserSession(ctx, testSession).Return(testUser.Id, nil)
				userRepo.EXPECT().GetUserByUId(ctx, testUser.Id).Return(testUser, nil)
			},
			session:     testSession,
			expectedErr: nil,
		},
		{
			name: "Session not found",
			mockSetup: func() {
				sessionRepo.EXPECT().LookupUserSession(ctx, testSession).Return(uuid.Nil, usecase.ErrNotFound)
			},
			session:     testSession,
			expectedErr: errors.New("a.sessionRepo.LookupUserSession: not found"),
		},
		{
			name: "User not found",
			mockSetup: func() {
				sessionRepo.EXPECT().LookupUserSession(ctx, testSession).Return(testUser.Id, nil)
				userRepo.EXPECT().GetUserByUId(ctx, testUser.Id).Return(models.User{}, usecase.ErrNotFound)
			},
			session:     testSession,
			expectedErr: errors.New("a.userRepo.GetUserByUId: not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := authService.LookupUserSession(ctx, tt.session)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testUser, user)
			}
		})
	}
}

func TestAuthService_DeleteUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	profileRepo := mocks.NewMockProfileRepository(ctrl)

	authService := usecase.NewAuthService(userRepo, sessionRepo, profileRepo)

	ctx := context.Background()
	sessionId := uuid.New().String()

	tests := []struct {
		name        string
		mockSetup   func()
		sessionId   string
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				sessionRepo.EXPECT().DeleteSession(ctx, sessionId).Return(nil)
			},
			sessionId:   sessionId,
			expectedErr: nil,
		},
		{
			name: "Error deleting session",
			mockSetup: func() {
				sessionRepo.EXPECT().DeleteSession(ctx, sessionId).Return(errors.New("delete error"))
			},
			sessionId:   sessionId,
			expectedErr: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := authService.DeleteUserSession(ctx, tt.sessionId)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_GetUserByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	profileRepo := mocks.NewMockProfileRepository(ctrl)

	authService := usecase.NewAuthService(userRepo, sessionRepo, profileRepo)

	ctx := context.Background()
	testUser := models.User{Id: uuid.New(), Username: "testuser"}
	username := "testuser"

	tests := []struct {
		name        string
		mockSetup   func()
		username    string
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				userRepo.EXPECT().GetUserByUsername(ctx, username).Return(testUser, nil)
			},
			username:    username,
			expectedErr: nil,
		},
		{
			name: "User not found",
			mockSetup: func() {
				userRepo.EXPECT().GetUserByUsername(ctx, username).Return(models.User{}, usecase.ErrNotFound)
			},
			username:    username,
			expectedErr: errors.New("a.userRepo.GetUserByUId: not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := authService.GetUserByUsername(ctx, tt.username)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testUser, user)
			}
		})
	}
}
