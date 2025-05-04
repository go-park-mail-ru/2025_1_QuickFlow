package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"strings"
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

func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "B"):
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	default:
		// по умолчанию — в байтах
	}

	num, err := strconv.ParseFloat(strings.TrimSpace(sizeStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}

	return int64(num * float64(multiplier)), nil
}
