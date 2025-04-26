package usecase

import (
	"context"
	"errors"
	"fmt"
	models2 "quickflow/monolith/internal/models"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type UserRepository interface {
	SaveUser(ctx context.Context, user models2.User) (uuid.UUID, error)
	GetUser(ctx context.Context, authData models2.LoginData) (models2.User, error)
	GetUserByUsername(ctx context.Context, username string) (models2.User, error)
	GetUserByUId(ctx context.Context, uid uuid.UUID) (models2.User, error)
	IsExists(ctx context.Context, login string) (bool, error)

	SearchSimilar(ctx context.Context, toSearch string, postsCount uint) ([]models2.PublicUserInfo, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, userId uuid.UUID, session models2.Session) error
	LookupUserSession(ctx context.Context, session models2.Session) (uuid.UUID, error)
	IsExists(ctx context.Context, sessionId uuid.UUID) (bool, error)
	DeleteSession(ctx context.Context, sessionId string) error
}

type AuthService struct {
	userRepo    UserRepository
	profileRepo ProfileRepository
	sessionRepo SessionRepository
}

// NewAuthService creates new auth service.
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, profileRepo ProfileRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		profileRepo: profileRepo,
	}
}

// CreateUser creates new user.
func (a *AuthService) CreateUser(ctx context.Context, user models2.User, profile models2.Profile) (uuid.UUID, models2.Session, error) {
	var err error
	if user, err = models2.CreateUser(user); err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("models.CreateUser: %w", err)
	}

	exists, err := a.userRepo.IsExists(ctx, user.Username)
	if err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("a.userRepo.IsExists: %w", err)
	}

	if exists {
		return uuid.Nil, models2.Session{}, ErrAlreadyExists
	}

	userId, err := a.userRepo.SaveUser(ctx, user)
	if err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("a.userRepo.SaveUser: %w", err)
	}
	profile.UserId = userId

	if err = a.profileRepo.SaveProfile(ctx, profile); err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("a.profileRepo.SaveProfile: %w", err)
	}

	session := models2.CreateSession()
	exists, err = a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = models2.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, userId, session); err != nil {
		return uuid.Nil, models2.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return userId, session, nil
}

// AuthUser checks if user exists and creates session.
func (a *AuthService) AuthUser(ctx context.Context, authData models2.LoginData) (models2.Session, error) {
	user, err := a.userRepo.GetUser(ctx, authData)
	if err != nil {
		return models2.Session{}, fmt.Errorf("a.userRepo.GetUser: %w", err)
	}

	session := models2.CreateSession()
	exists, err := a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return models2.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = models2.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, user.Id, session); err != nil {
		return models2.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return session, nil
}

func (a *AuthService) GetUserByUsername(ctx context.Context, username string) (models2.User, error) {
	user, err := a.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return models2.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// LookupUserSession returns user by session.
func (a *AuthService) LookupUserSession(ctx context.Context, session models2.Session) (models2.User, error) {
	userID, err := a.sessionRepo.LookupUserSession(ctx, session)
	if err != nil {
		return models2.User{}, fmt.Errorf("a.sessionRepo.LookupUserSession: %w", err)
	}

	user, err := a.userRepo.GetUserByUId(ctx, userID)
	if err != nil {
		return models2.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// DeleteUserSession deletes user session.
func (a *AuthService) DeleteUserSession(ctx context.Context, sessionId string) error {
	return a.sessionRepo.DeleteSession(ctx, sessionId)
}
