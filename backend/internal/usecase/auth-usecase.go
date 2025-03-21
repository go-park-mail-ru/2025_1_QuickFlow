package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"

	"quickflow/internal/models"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user models.User) (uuid.UUID, error)
	GetUser(ctx context.Context, authData models.LoginData) (models.User, error)
	GetUserByUId(ctx context.Context, uid uuid.UUID) (models.User, error)
	IsExists(ctx context.Context, login string) (bool, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, userId uuid.UUID, session models.Session) error
	LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error)
	IsExists(ctx context.Context, sessionId uuid.UUID) (bool, error)
	DeleteSession(ctx context.Context, sessionId string) error
}

type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
}

// NewAuthService creates new auth service.
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateUser creates new user.
func (a *AuthService) CreateUser(ctx context.Context, user models.User) (uuid.UUID, models.Session, error) {
	var err error
	if user, err = models.CreateUser(user); err != nil {
		return uuid.Nil, models.Session{}, err
	}

	exists, err := a.userRepo.IsExists(ctx, user.Login)
	if err != nil {
		return uuid.Nil, models.Session{}, fmt.Errorf("p.repo.IsExists: %w", err)
	}

	if exists {
		return uuid.Nil, models.Session{}, errors.New("user already exists")
	}

	userId, err := a.userRepo.SaveUser(ctx, user)
	if err != nil {
		return uuid.Nil, models.Session{}, err
	}

	session := models.CreateSession()
	exists, err = a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return uuid.Nil, models.Session{}, fmt.Errorf("p.repo.IsExists: %w", err)
	}

	if exists {
		session = models.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, userId, session); err != nil {
		return uuid.Nil, models.Session{}, err
	}

	return userId, session, nil
}

// GetUser checks if user exists and creates session.
func (a *AuthService) GetUser(ctx context.Context, authData models.LoginData) (models.Session, error) {
	user, err := a.userRepo.GetUser(ctx, authData)
	if err != nil {
		return models.Session{}, err
	}

	session := models.CreateSession()
	exists, err := a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return models.Session{}, fmt.Errorf("p.repo.IsExists: %w", err)
	}

	if exists {
		session = models.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, user.Id, session); err != nil {
		return models.Session{}, err
	}

	return session, nil
}

// LookupUserSession returns user by session.
func (a *AuthService) LookupUserSession(ctx context.Context, session models.Session) (models.User, error) {
	userID, err := a.sessionRepo.LookupUserSession(ctx, session)
	if err != nil {
		return models.User{}, fmt.Errorf("p.repo.LookupUserSession: %w", err)
	}

	user, err := a.userRepo.GetUserByUId(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("p.repo.GetUserByUId: %w", err)
	}

	return user, nil
}

func (a *AuthService) DeleteUserSession(ctx context.Context, sessionId string) error {
	return a.sessionRepo.DeleteSession(ctx, sessionId)
}
