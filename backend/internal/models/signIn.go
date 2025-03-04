package models

type AuthForm struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
