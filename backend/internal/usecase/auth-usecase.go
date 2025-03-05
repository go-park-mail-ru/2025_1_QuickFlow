package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type UserRepository interface {
	SaveUser(user models.User) (uuid.UUID, error)
	GetUser(authData models.AuthForm) (models.User, error)
	GetUserByUId(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type SessionRepository interface {
	CreateSession(userId uuid.UUID) models.Session
	LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error)
}

type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
}

func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (a *AuthService) CreateUser(user models.User) (uuid.UUID, models.Session, error) {
	userId, err := a.userRepo.SaveUser(user)
	if err != nil {
		return uuid.Nil, models.Session{}, err
	}

	session := a.sessionRepo.CreateSession(userId)

	return userId, session, nil
}

func (a *AuthService) GetUser(authData models.AuthForm) (models.Session, error) {
	user, err := a.userRepo.GetUser(authData)
	if err != nil {
		return models.Session{}, err
	}

	session := a.sessionRepo.CreateSession(user.Id)

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
