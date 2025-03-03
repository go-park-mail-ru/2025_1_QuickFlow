package models

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	//"regexp"
	//"time"
	"unicode"
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

type AuthForm struct {
	Login    string
	Password string
}

const randLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const randDigits = "0123456789"

func validateLogin(login string) bool {
	return true
}

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasSpecial, hasDigit bool
	for _, char := range password {
		switch {

		case unicode.IsUpper(char):
			hasUpper = true

		case unicode.IsLower(char):
			hasLower = true

		case unicode.IsDigit(char):
			hasDigit = true

		case unicode.IsPunct(char):
			hasSpecial = true

		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

//func validatePhone(phone string) bool {
//	re := regexp.MustCompile("^((\\+7|7|8)+([0-9]){10})$")
//	return re.MatchString(phone)
//}

//func validateEmail(email string) bool {
//	re := regexp.MustCompile("[a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\\.[a-zA-Z0-9_-]")
//	return re.MatchString(email)
//
//}

func Validate(user User) error {
	switch {

	case !validateLogin(user.Login):
		return fmt.Errorf("invalid login %s", user.Login)

	case !validatePassword(user.Password):
		return fmt.Errorf("invalid password %s", user.Password)
	}

	return nil
}

func CheckPassword(password string, user User) bool {
	passwordCheck := sha256.Sum256([]byte(password + user.Salt))
	if string(passwordCheck[:]) == user.Password {
		return true
	}
	return false
}

func hashPassword(password string, salt string) string {
	data := password + salt
	hash := sha256.Sum256([]byte(data))

	return string(hash[:])

}

func genSalt() string {
	res := make([]byte, 10)
	for i := 0; i < 10; i++ {
		res[i] = randLetters[rand.Intn(len(randLetters))]
	}

	return string(res)
}

func genLogin() string {
	res := make([]byte, 9)
	for i := 0; i < 9; i++ {
		res[i] = randDigits[rand.Intn(len(randDigits))]
	}

	return "id" + string(res)
}

func CreateUser(user User) User {
	id := uuid.New()
	salt := genSalt()
	login := genLogin()
	hashedPassword := hashPassword(user.Password, salt)

	newUser := User{
		Id:          id,
		Login:       login,
		Name:        user.Name,
		Surname:     user.Surname,
		Sex:         user.Sex,
		DateOfBirth: user.DateOfBirth,
		Password:    hashedPassword,
		Salt:        salt,
	}

	return newUser

}
