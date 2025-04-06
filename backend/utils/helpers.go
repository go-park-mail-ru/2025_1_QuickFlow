package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	mathrand "math/rand"
)

const AcceptableSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_/!@#$%^&*(),.?\":{}|<>"

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func HashPassword(password, salt string) string {
	data := password + salt
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func GenSalt() string {
	res := make([]byte, 10)
	for i := 0; i < 10; i++ {
		res[i] = AcceptableSymbols[mathrand.Intn(len(AcceptableSymbols))]
	}

	return string(res)
}
