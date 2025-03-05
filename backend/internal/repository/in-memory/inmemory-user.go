package in_memory

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils"
)

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]models.User
}

// NewInMemoryUserRepository creates new storage instance.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]models.User),
	}
}

// SaveUser saves user to the repository.
func (i *InMemoryUserRepository) SaveUser(user models.User) (uuid.UUID, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.users[user.Login] = user

	return user.Id, nil
}

// GetUser returns user by login and password.
func (i *InMemoryUserRepository) GetUser(authData models.AuthForm) (models.User, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	user, exists := i.users[authData.Login]

	switch {

	case !exists:
		return models.User{}, errors.New("user not found")

	case !utils.CheckPassword(authData.Password, user.Password, user.Salt):
		return models.User{}, errors.New("incorrect login or password")
	}

	return user, nil
}

// GetUserByUId returns user by id.
func (i *InMemoryUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models.User, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, user := range i.users {
		if user.Id == userId {
			return user, nil
		}
	}

	return models.User{}, errors.New("user not found")
}

func (i *InMemoryUserRepository) GetUsers() map[string]models.User {
	return i.users
}
