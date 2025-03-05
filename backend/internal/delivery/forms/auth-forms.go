package forms

import "quickflow/internal/models"

type SignUpForm struct {
	Login       string     `json:"login"`
	Password    string     `json:"password"`
	Name        string     `json:"name"`
	Surname     string     `json:"surname"`
	Sex         models.Sex `json:"sex"`
	DateOfBirth string     `json:"date_of_birth"`
}

type AuthForm struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
