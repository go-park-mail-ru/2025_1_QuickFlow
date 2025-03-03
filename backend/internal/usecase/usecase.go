package usecase

import (
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type repository interface {
	SaveUser(user models.User) (uuid.UUID, models.Session, error)
	GetUser(authData models.AuthForm) (models.Session, error)
}

type Processor struct {
	repo repository
}

func NewProcessor(repo repository) *Processor {
	return &Processor{repo: repo}
}

func (p *Processor) CreateUser(user models.User) (uuid.UUID, models.Session, error) {
	id, session, err := p.repo.SaveUser(user)
	if err != nil {
		return uuid.Nil, models.Session{}, err
	}

	return id, session, nil
}

func (p *Processor) GetUser(authData models.AuthForm) (models.Session, error) {
	session, err := p.repo.GetUser(authData)

	return session, err
}
