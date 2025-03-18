package usecase

import (
    "context"
    "testing"

    "github.com/golang/mock/gomock"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "quickflow/internal/models"
    mock_usecase "quickflow/internal/usecase/mocks"
)

func TestCreateUser(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUserRepo := mock_usecase.NewMockUserRepository(ctrl)
    mockSessionRepo := mock_usecase.NewMockSessionRepository(ctrl)
    authService := NewAuthService(mockUserRepo, mockSessionRepo)

    ctx := context.Background()
    user := models.User{Login: "testuser", Password: "securepassword"}
    userID := uuid.New()

    mockUserRepo.EXPECT().IsExists(ctx, user.Login).Return(false, nil).Times(1)
    mockUserRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(userID, nil).Times(1)
    mockSessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil).Times(1)
    mockSessionRepo.EXPECT().SaveSession(ctx, userID, gomock.Any()).Return(nil).Times(1)

    createdUserID, createdSession, err := authService.CreateUser(ctx, user)

    assert.NoError(t, err)
    assert.Equal(t, userID, createdUserID)
    assert.NotEqual(t, uuid.Nil, createdSession.SessionId)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUserRepo := mock_usecase.NewMockUserRepository(ctrl)
    mockSessionRepo := mock_usecase.NewMockSessionRepository(ctrl)
    authService := NewAuthService(mockUserRepo, mockSessionRepo)

    ctx := context.Background()
    user := models.User{Login: "existinguser", Password: "securepassword"}

    mockUserRepo.EXPECT().IsExists(ctx, user.Login).Return(true, nil).Times(1)

    _, _, err := authService.CreateUser(ctx, user)

    assert.Error(t, err)
    assert.Equal(t, "user already exists", err.Error())
}

func TestGetUser(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUserRepo := mock_usecase.NewMockUserRepository(ctrl)
    mockSessionRepo := mock_usecase.NewMockSessionRepository(ctrl)
    authService := NewAuthService(mockUserRepo, mockSessionRepo)

    ctx := context.Background()
    loginData := models.LoginData{Login: "testuser", Password: "securepassword"}
    user := models.User{Id: uuid.New(), Login: loginData.Login, Password: loginData.Password}

    mockUserRepo.EXPECT().GetUser(ctx, loginData).Return(user, nil).Times(1)
    mockSessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil).Times(1)
    mockSessionRepo.EXPECT().SaveSession(ctx, user.Id, gomock.Any()).Return(nil).Times(1)

    createdSession, err := authService.GetUser(ctx, loginData)

    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, createdSession.SessionId)
}

func TestLookupUserSession(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUserRepo := mock_usecase.NewMockUserRepository(ctrl)
    mockSessionRepo := mock_usecase.NewMockSessionRepository(ctrl)
    authService := NewAuthService(mockUserRepo, mockSessionRepo)

    ctx := context.Background()
    session := models.CreateSession()
    userID := uuid.New()
    user := models.User{Id: userID, Login: "testuser"}

    mockSessionRepo.EXPECT().LookupUserSession(ctx, session).Return(userID, nil).Times(1)
    mockUserRepo.EXPECT().GetUserByUId(ctx, userID).Return(user, nil).Times(1)

    foundUser, err := authService.LookupUserSession(ctx, session)

    assert.NoError(t, err)
    assert.Equal(t, user, foundUser)
}

func TestDeleteUserSession(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSessionRepo := mock_usecase.NewMockSessionRepository(ctrl)
    authService := NewAuthService(nil, mockSessionRepo)

    ctx := context.Background()
    sessionID := uuid.New().String()

    mockSessionRepo.EXPECT().DeleteSession(ctx, sessionID).Return(nil).Times(1)

    err := authService.DeleteUserSession(ctx, sessionID)

    assert.NoError(t, err)
}
