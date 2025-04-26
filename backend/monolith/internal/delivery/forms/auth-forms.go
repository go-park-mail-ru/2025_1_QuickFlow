package forms

import (
	"quickflow/monolith/internal/models"
)

type SignUpForm struct {
	Login       string     `json:"username"`
	Password    string     `json:"password"`
	Name        string     `json:"firstname"`
	Surname     string     `json:"lastname"`
	Sex         models.Sex `json:"sex"`
	DateOfBirth string     `json:"birth_date"`
}

type AuthForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}
