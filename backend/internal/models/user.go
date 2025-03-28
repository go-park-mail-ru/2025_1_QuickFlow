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
    Id       uuid.UUID
    Login    string
    Password string
    Salt     string
}

// CreateUser creates new user.
func CreateUser(user User) (User, error) {
    id := uuid.New()
    salt := validation.GenSalt()
    hashedPassword := validation.HashPassword(user.Password, salt)

    newUser := User{
        Id:       id,
        Login:    user.Login,
        Password: hashedPassword,
        Salt:     salt,
    }

    return newUser, nil
}
