package models

import (
	"time"

	"github.com/google/uuid"

	"quickflow/utils"
)

type Sex int

const (
	MALE = iota
	FEMALE
)

type User struct {
	Id       uuid.UUID
	Username string
	Password string
	Salt     string
	LastSeen time.Time
}

// CreateUser creates new user.
func CreateUser(user User) (User, error) {
	id := uuid.New()
	salt := utils.GenSalt()
	hashedPassword := utils.HashPassword(user.Password, salt)

	newUser := User{
		Id:       id,
		Username: user.Username,
		Password: hashedPassword,
		Salt:     salt,
	}

	return newUser, nil
}
