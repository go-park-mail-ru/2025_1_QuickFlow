package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	shared_models "quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/utils/validation"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user shared_models.User) (uuid.UUID, error)
	GetUser(ctx context.Context, authData shared_models.LoginData) (shared_models.User, error)
	GetUserByUsername(ctx context.Context, username string) (shared_models.User, error)
	GetUserByUId(ctx context.Context, uid uuid.UUID) (shared_models.User, error)
	IsExists(ctx context.Context, login string) (bool, error)

	SearchSimilar(ctx context.Context, toSearch string, postsCount uint) ([]shared_models.PublicUserInfo, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, userId uuid.UUID, session shared_models.Session) error
	LookupUserSession(ctx context.Context, session shared_models.Session) (uuid.UUID, error)
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
func (a *AuthService) CreateUser(ctx context.Context, user shared_models.User, profile shared_models.Profile) (uuid.UUID, shared_models.Session, error) {
	var err error
	if user, err = shared_models.CreateUser(user); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("shared_models.CreateUser: %w", err)
	}

	exists, err := a.userRepo.IsExists(ctx, user.Username)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.userRepo.IsExists: %w", err)
	}

	if exists {
		return uuid.Nil, shared_models.Session{}, user_errors.ErrAlreadyExists
	}

	// validation
	if err = validation.ValidateUser(user.Username, user.Password); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("%w: validation.ValidateUser: %w", user_errors.ErrUserValidation, err)
	}
	if err = validation.ValidateProfile(profile.BasicInfo.Name, profile.BasicInfo.Surname); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("%w: validation.ValidateProfile: %w", user_errors.ErrProfileValidation, err)
	}

	userId, err := a.userRepo.SaveUser(ctx, user)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.userRepo.SaveUser: %w", err)
	}
	profile.UserId = userId

	if err = a.profileRepo.SaveProfile(ctx, profile); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.profileRepo.SaveProfile: %w", err)
	}

	session := shared_models.CreateSession()
	exists, err = a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = shared_models.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, userId, session); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return userId, session, nil
}

// AuthUser checks if user exists and creates session.
func (a *AuthService) AuthUser(ctx context.Context, authData shared_models.LoginData) (shared_models.Session, error) {
	user, err := a.userRepo.GetUser(ctx, authData)
	if err != nil {
		return shared_models.Session{}, fmt.Errorf("a.userRepo.GetUser: %w", err)
	}

	session := shared_models.CreateSession()
	exists, err := a.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return shared_models.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = shared_models.CreateSession()
	}

	if err = a.sessionRepo.SaveSession(ctx, user.Id, session); err != nil {
		return shared_models.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return session, nil
}

func (a *AuthService) GetUserByUsername(ctx context.Context, username string) (shared_models.User, error) {
	user, err := a.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// LookupUserSession returns user by session.
func (a *AuthService) LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error) {
	userID, err := a.sessionRepo.LookupUserSession(ctx, session)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.sessionRepo.LookupUserSession: %w", err)
	}

	user, err := a.userRepo.GetUserByUId(ctx, userID)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// DeleteUserSession deletes user session.
func (a *AuthService) DeleteUserSession(ctx context.Context, sessionId string) error {
	return a.sessionRepo.DeleteSession(ctx, sessionId)
}

func (a *AuthService) GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error) {
	user, err := a.userRepo.GetUserByUId(ctx, userId)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}
