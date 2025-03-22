package models

import (
	"github.com/google/uuid"
	"quickflow/utils/validation"
)

type Sex int

const (
	MALE = iota
	FEMALE
)

type User struct {
	Id          uuid.UUID
	Login       string
	Name        string
	Surname     string
	Sex         Sex
	DateOfBirth string
	Password    string
	Salt        string
}

// CreateUser creates new user.
func CreateUser(user User) (User, error) {
	id := uuid.New()
	salt := validation.GenSalt()
	hashedPassword := validation.HashPassword(user.Password, salt)

	newUser := User{
		Id:          id,
		Login:       user.Login,
		Name:        user.Name,
		Surname:     user.Surname,
		Sex:         user.Sex,
		DateOfBirth: user.DateOfBirth,
		Password:    hashedPassword,
		Salt:        salt,
	}

	return newUser, nil
}
