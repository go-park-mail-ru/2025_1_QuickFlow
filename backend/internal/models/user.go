package models

import (
	"fmt"

	"github.com/google/uuid"

	"quickflow/utils"
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
func CreateUser(user User, users map[string]User) (User, error) {
	if _, ok := users[user.Login]; ok {
		return User{}, fmt.Errorf("this login already exists")
	}

	id := uuid.New()
	salt := utils.GenSalt()
	hashedPassword := utils.HashPassword(user.Password, salt)

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
