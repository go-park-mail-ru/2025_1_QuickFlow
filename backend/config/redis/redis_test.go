package redis_config

import (
	"os"
	"testing"
)

func TestNewRedisConfig_WithEnvVar(t *testing.T) {
	// Устанавливаем переменную окружения
	os.Setenv("REDIS_URL", "redis://custom-redis:6379")
	defer os.Unsetenv("REDIS_URL") // Очищаем после теста

	// Создаем новый объект конфигурации
	redisConfig := NewRedisConfig()

	// Проверяем, что возвращается правильный URL
	expectedURL := "redis://custom-redis:6379"
	if redisConfig.GetURL() != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, redisConfig.GetURL())
	}
}

func TestNewRedisConfig_WithoutEnvVar(t *testing.T) {
	// Убираем переменную окружения, если она установлена
	os.Unsetenv("REDIS_URL")

	// Создаем новый объект конфигурации
	redisConfig := NewRedisConfig()

	// Проверяем, что возвращается URL по умолчанию
	expectedURL := backupURL
	if redisConfig.GetURL() != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, redisConfig.GetURL())
	}
}
