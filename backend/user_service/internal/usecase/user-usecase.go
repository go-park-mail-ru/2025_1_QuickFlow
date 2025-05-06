package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	shared_models "quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/models"
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

type UserUseCase struct {
	userRepo    UserRepository
	profileRepo ProfileRepository
	sessionRepo SessionRepository
}

// NewUserUseCase creates new auth service.
func NewUserUseCase(userRepo UserRepository, sessionRepo SessionRepository, profileRepo ProfileRepository) *UserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		profileRepo: profileRepo,
	}
}

// CreateUser creates new user.
func (u *UserUseCase) CreateUser(ctx context.Context, user shared_models.User, profile shared_models.Profile) (uuid.UUID, shared_models.Session, error) {
	var err error
	// validation
	if err = validation.ValidateUser(user.Username, user.Password); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("%w: validation.ValidateUser: %w", user_errors.ErrUserValidation, err)
	}

	if profile.BasicInfo == nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("%w: profile.BasicInfo is nil", user_errors.ErrProfileValidation)
	}

	if err = validation.ValidateProfile(profile.BasicInfo.Name, profile.BasicInfo.Surname); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("%w: validation.ValidateProfile: %w", user_errors.ErrProfileValidation, err)
	}

	if user, err = models.CreateUser(user); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("shared_models.CreateUser: %w", err)
	}

	exists, err := u.userRepo.IsExists(ctx, user.Username)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.userRepo.IsExists: %w", err)
	}

	if exists {
		return uuid.Nil, shared_models.Session{}, user_errors.ErrAlreadyExists
	}

	userId, err := u.userRepo.SaveUser(ctx, user)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.userRepo.SaveUser: %w", err)
	}
	profile.UserId = userId

	if err = u.profileRepo.SaveProfile(ctx, profile); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.profileRepo.SaveProfile: %w", err)
	}

	session := shared_models.CreateSession()
	exists, err = u.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = shared_models.CreateSession()
	}

	if err = u.sessionRepo.SaveSession(ctx, userId, session); err != nil {
		return uuid.Nil, shared_models.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return userId, session, nil
}

// AuthUser checks if user exists and creates session.
func (u *UserUseCase) AuthUser(ctx context.Context, authData shared_models.LoginData) (shared_models.Session, error) {
	user, err := u.userRepo.GetUser(ctx, authData)
	if err != nil {
		return shared_models.Session{}, fmt.Errorf("a.userRepo.GetUser: %w", err)
	}

	session := shared_models.CreateSession()
	exists, err := u.sessionRepo.IsExists(ctx, session.SessionId)
	if err != nil {
		return shared_models.Session{}, fmt.Errorf("a.sessionRepo.IsExists: %w", err)
	}

	if exists {
		session = shared_models.CreateSession()
	}

	if err = u.sessionRepo.SaveSession(ctx, user.Id, session); err != nil {
		return shared_models.Session{}, fmt.Errorf("a.sessionRepo.SaveSession: %w", err)
	}

	return session, nil
}

func (u *UserUseCase) GetUserByUsername(ctx context.Context, username string) (shared_models.User, error) {
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// LookupUserSession returns user by session.
func (u *UserUseCase) LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error) {
	userID, err := u.sessionRepo.LookupUserSession(ctx, session)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.sessionRepo.LookupUserSession: %w", err)
	}

	user, err := u.userRepo.GetUserByUId(ctx, userID)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

// DeleteUserSession deletes user session.
func (u *UserUseCase) DeleteUserSession(ctx context.Context, sessionId string) error {
	return u.sessionRepo.DeleteSession(ctx, sessionId)
}

func (u *UserUseCase) GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error) {
	user, err := u.userRepo.GetUserByUId(ctx, userId)
	if err != nil {
		return shared_models.User{}, fmt.Errorf("a.userRepo.GetUserByUId: %w", err)
	}

	return user, nil
}

func (u *UserUseCase) SearchSimilarUser(ctx context.Context, toSearch string, postsCount uint) ([]shared_models.PublicUserInfo, error) {
	users, err := u.userRepo.SearchSimilar(ctx, toSearch, postsCount)
	if err != nil {
		return nil, err
	}

	return users, nil
}
