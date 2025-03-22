package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"unicode"
)

const (
	randSpecialSymbols = "_/!@#$%^&*(),.?\":{}|<>"
	acceptableSymbols  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_/!@#$%^&*(),.?\":{}|<>"
	acceptableLogin    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ._123456789"
)

func validateLogin(login string) bool {
	switch {
	case len(login) < 1 || len(login) > 20:
		return false

	case strings.ContainsRune(login, ' '):
		return false

	case rune(login[0]) == '.':
		return false

	case rune(login[0]) == '_':
		return false
	}

	for _, char := range login {
		if !strings.ContainsRune(acceptableLogin, char) {
			return false
		}
	}

	return true
}

func validatePassword(password string) bool {
	if len(password) < 8 || len(password) > 32 {
		return false
	}

	if strings.ContainsRune(password, ' ') {
		return false
	}

	var hasUpper, hasLower, hasSpecial, hasDigit bool

	for _, char := range password {
		switch {

		case !strings.ContainsRune(acceptableSymbols, char):
			return false

		case unicode.IsUpper(char):
			hasUpper = true

		case unicode.IsLower(char):
			hasLower = true

		case unicode.IsDigit(char):
			hasDigit = true

		case strings.ContainsRune(randSpecialSymbols, char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateCreds(str string) bool {
	runeName := []rune(str)
	if len(str) < 2 || len(str) > 25 {
		return false
	}

	if runeName[len(runeName)-1] == '-' || runeName[0] == '-' {
		return false
	}

	underline := strings.Count(str, "_")
	if underline > 1 {
		return false
	}

	splitedStr := strings.Split(str, "-")

	if len(splitedStr) > 2 {
		return false
	}

	for _, chunk := range splitedStr {
		for _, char := range chunk {
			if !unicode.IsLetter(char) {
				return false
			}
		}
	}

	return true
}

func Validate(login string, password string, firstName string, lastName string) error {
	switch {

	case !validateLogin(login):
		return errors.New("invalid login")

	case !validatePassword(password):
		return errors.New("invalid password")

	case !validateCreds(firstName):
		return errors.New("invalid first name")

	case !validateCreds(lastName):
		return errors.New("invalid last name")

	}

	return nil
}

func CheckPassword(password string, userPassword string, userSalt string) bool {
	passwordCheck := sha256.Sum256([]byte(password + userSalt))
	if hex.EncodeToString(passwordCheck[:]) == userPassword {
		return true
	}
	return false
}

func HashPassword(password string, salt string) string {
	data := password + salt
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func GenSalt() string {
	res := make([]byte, 10)
	for i := 0; i < 10; i++ {
		res[i] = acceptableSymbols[rand.Intn(len(acceptableSymbols))]
	}

	return string(res)
}
