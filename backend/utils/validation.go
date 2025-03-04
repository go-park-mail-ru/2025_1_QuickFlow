package utils

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

const randLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789_"
const randDigits = "0123456789"

func validateLogin(login string) bool {
	if len(login) < 1 || len(login) > 20 {
		return false
	}

	for _, digit := range randDigits {
		if login[1] == byte(digit) {
			return false
		}
	}

	for _, ch := range login {
		if !strings.ContainsRune(randLetters, ch) {
			return false
		}
	}

	return true

}

func validatePassword(password string) bool {
	if len(password) < 8 || len(password) > 32 {
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

func Validate(login string, password string) error {
	switch {

	case !validateLogin(login):
		return fmt.Errorf("invalid login")

	case !validatePassword(password):
		return fmt.Errorf("invalid password")
	}

	return nil
}

func CheckPassword(password string, userPassword string, userSalt string) bool {
	passwordCheck := sha256.Sum256([]byte(password + userSalt))
	if string(passwordCheck[:]) == userPassword {
		return true
	}
	return false
}

func HashPassword(password string, salt string) string {
	data := password + salt
	hash := sha256.Sum256([]byte(data))

	return string(hash[:])

}

func GenSalt() string {
	res := make([]byte, 10)
	for i := 0; i < 10; i++ {
		res[i] = randLetters[rand.Intn(len(randLetters))]
	}

	return string(res)
}
