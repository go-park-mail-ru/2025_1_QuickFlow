package repository

import (
	"fmt"
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type InMemory struct {
	users    map[string]models.User
	sessions map[uuid.UUID]models.User
}

func NewInMemory() *InMemory {
	return &InMemory{
		users:    map[string]models.User{},
		sessions: map[uuid.UUID]models.User{},
	}
}

func (p *InMemory) SaveUser(user models.User) (uuid.UUID, models.Session, error) {
	session := models.CreateSession()
	user = models.CreateUser(user)
	p.sessions[session.SessionId] = user
	p.users[user.Login] = user
	fmt.Println(user.Login)

	return user.Id, session, nil
}

func (p *InMemory) GetUser(authData models.AuthForm) (models.Session, error) {
	user, exists := p.users[authData.Login]

	if !exists || !models.CheckPassword(authData.Password, user) {
		return models.Session{}, fmt.Errorf("incorrect login or password")
	}

	newSession := models.CreateSession()
	p.sessions[newSession.SessionId] = user

	return newSession, nil

}
