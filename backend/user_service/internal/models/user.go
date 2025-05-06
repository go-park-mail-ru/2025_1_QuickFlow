package models

import (
	"github.com/google/uuid"

	"quickflow/shared/models"
	"quickflow/user_service/utils"
)

// CreateUser creates new user.
func CreateUser(user models.User) (models.User, error) {
	id := uuid.New()
	salt := utils.GenSalt()
	hashedPassword := utils.HashPassword(user.Password, salt)

	newUser := models.User{
		Id:       id,
		Username: user.Username,
		Password: hashedPassword,
		Salt:     salt,
	}

	return newUser, nil
}
