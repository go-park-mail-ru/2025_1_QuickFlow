package utils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken(t *testing.T) {
	t.Run("Test GenerateCSRFToken", func(t *testing.T) {
		token, err := GenerateCSRFToken()

		// Проверка на ошибку
		assert.NoError(t, err)

		// Проверка на ненулевое значение токена
		assert.NotEmpty(t, token)

		// Проверка на правильный формат (base64 строка)
		_, err = base64.StdEncoding.DecodeString(token)
		assert.NoError(t, err, "Token should be a valid base64 encoded string")
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("Test HashPassword", func(t *testing.T) {
		password := "testPassword"
		salt := "testSalt"

		// Хешируем пароль с солью
		hashedPassword := HashPassword(password, salt)

		// Проверка, что результат не пустой
		assert.NotEmpty(t, hashedPassword)

		// Проверка, что хеш не совпадает с исходным паролем
		// (хеширование пароля делает его уникальным, даже если пароль одинаковый)
		assert.NotEqual(t, password+salt, hashedPassword)

		// Проверка, что хеш не совпадает с другим паролем с такой же солью
		anotherPassword := "anotherPassword"
		anotherHashedPassword := HashPassword(anotherPassword, salt)
		assert.NotEqual(t, hashedPassword, anotherHashedPassword)
	})
}

func TestGenSalt(t *testing.T) {
	t.Run("Test GenSalt", func(t *testing.T) {
		// Генерация соли
		salt := GenSalt()

		// Проверка, что соль не пустая
		assert.NotEmpty(t, salt)

		// Проверка длины соли
		assert.Len(t, salt, 10)

		// Проверка, что соль состоит только из допустимых символов
		for _, ch := range salt {
			assert.Contains(t, AcceptableSymbols, string(ch), "Salt contains invalid character")
		}
	})
}
