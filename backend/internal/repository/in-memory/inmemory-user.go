package in_memory

import (
	"context"
	"errors"
	"github.com/google/uuid"

	"quickflow/internal/models"
	tsmap "quickflow/pkg/thread-safe-map"
	"quickflow/utils"
)

type InMemoryUserRepository struct {
	users tsmap.ThreadSafeMap[string, models.User]
}

// NewInMemoryUserRepository creates new storage instance.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: *tsmap.NewThreadSafeMap[string, models.User](),
	}
}

// SaveUser saves user to the repository.
func (i *InMemoryUserRepository) SaveUser(_ context.Context, user models.User) (uuid.UUID, error) {
	i.users.Set(user.Login, user)

	return user.Id, nil
}

// GetUser returns user by login and password.
func (i *InMemoryUserRepository) GetUser(_ context.Context, loginData models.LoginData) (models.User, error) {
	user, exists := i.users.Get(loginData.Login)

	switch {

	case !exists:
		return models.User{}, errors.New("user not found")

	case !utils.CheckPassword(loginData.Password, user.Password, user.Salt):
		return models.User{}, errors.New("incorrect login or password")
	}

	return user, nil
}

// GetUserByUId returns user by id.
func (i *InMemoryUserRepository) GetUserByUId(_ context.Context, userId uuid.UUID) (models.User, error) {
	var user models.User

	for user = range i.users.GetValues() {
		if user.Id == userId {
			return user, nil
		}
	}

	return models.User{}, errors.New("user not found")
}

func (i *InMemoryUserRepository) IsExists(_ context.Context, login string) bool {
	if _, exists := i.users.Get(login); exists {
		return true
	}

	return false
}
