package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"unicode"

	"quickflow/gateway/utils"
)

const (
	acceptableLogin = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ._0123456789"
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

	var hasUpper, hasLower, hasDigit bool

	for _, char := range password {
		switch {

		case !strings.ContainsRune(utils.AcceptableSymbols, char):
			return false

		case unicode.IsUpper(char):
			hasUpper = true

		case unicode.IsLower(char):
			hasLower = true

		case unicode.IsDigit(char):
			hasDigit = true

		}
	}

	return hasUpper && hasLower && hasDigit
}

func validateCreds(str string) bool {
	runeName := []rune(str)
	if len(runeName) < 2 || len(runeName) > 25 {
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

func ValidateUser(login, password string) error {
	switch {
	case !validateLogin(login):
		return errors.New("invalid login")

	case !validatePassword(password):
		return errors.New("invalid password")
	}
	return nil
}

func ValidateProfile(firstName, lastName string) error {
	switch {
	case !validateCreds(firstName):
		return errors.New("invalid first name")

	case !validateCreds(lastName):
		return errors.New("invalid last name")
	}

	return nil
}

func CheckPassword(password, userPassword, userSalt string) bool {
	passwordCheck := sha256.Sum256([]byte(password + userSalt))

	return hex.EncodeToString(passwordCheck[:]) == userPassword
}
