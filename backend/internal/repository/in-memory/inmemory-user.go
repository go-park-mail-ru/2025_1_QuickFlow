package in_memory

import (
	"context"
	"fmt"
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

	user, err := models.CreateUser(user, i.users)
	if err != nil {
		return uuid.Nil, err
	}

	i.users[user.Login] = user

	return user.Id, nil
}

// GetUser returns user by login and password.
func (i *InMemoryUserRepository) GetUser(authData models.AuthForm) (models.User, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	user, exists := i.users[authData.Login]
	if !exists || !utils.CheckPassword(authData.Password, user.Password, user.Salt) {
		return models.User{}, fmt.Errorf("incorrect login or password")
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

	return models.User{}, fmt.Errorf("user not found")
}
