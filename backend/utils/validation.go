package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

const lowerLetters = "abcdefghijklmnopqrstuvwxyz"
const upperLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randDigits = "0123456789"
const randSpecialSymbols = "/!@#$%^&*(),.?\":{}|<>"

func validateLogin(login string) bool {
	if len(login) < 1 || len(login) > 21 {
		return false
	}

	for _, digit := range randDigits {
		if login[1] == byte(digit) {
			return false
		}
	}

	for _, ch := range login {
		if !strings.ContainsRune(lowerLetters, ch) || !strings.ContainsRune(upperLetter, ch) {
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

		case !strings.ContainsRune(lowerLetters, char) || !strings.ContainsRune(upperLetter, char):
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
		res[i] = lowerLetters[rand.Intn(len(lowerLetters))]
	}

	return string(res)
}
