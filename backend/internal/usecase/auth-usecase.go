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
	GetUsers() map[string]models.User
}

type SessionRepository interface {
	SaveSession(userId uuid.UUID, session models.Session) models.Session
	LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error)
	GetSessions() map[uuid.UUID]uuid.UUID
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
	users := a.userRepo.GetUsers()

	user, err := models.CreateUser(user, users)
	if err != nil {
		return uuid.Nil, models.Session{}, err
	}

	userId, err := a.userRepo.SaveUser(user)
	if err != nil {
		return uuid.Nil, models.Session{}, err
	}

	sessions := a.sessionRepo.GetSessions()

	session := models.CreateSession(sessions)

	return userId, session, nil
}

func (a *AuthService) GetUser(authData models.AuthForm) (models.Session, error) {
	user, err := a.userRepo.GetUser(authData)
	if err != nil {
		return models.Session{}, err
	}

	sessions := a.sessionRepo.GetSessions()

	session := models.CreateSession(sessions)

	a.sessionRepo.SaveSession(user.Id, session)

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
