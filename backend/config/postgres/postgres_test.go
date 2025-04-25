package postgres_config

import (
	"os"
	"testing"
)

func TestNewPostgresConfig_WithEnvVar(t *testing.T) {
	// Устанавливаем переменную окружения
	os.Setenv("DATABASE_URL", "postgresql://custom-user:password@localhost:5432/custom_db")
	defer os.Unsetenv("DATABASE_URL") // Очищаем после теста

	// Создаем новый объект конфигурации
	postgresConfig := NewPostgresConfig()

	// Проверяем, что возвращается правильный URL
	expectedURL := "postgresql://custom-user:password@localhost:5432/custom_db"
	if postgresConfig.GetURL() != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, postgresConfig.GetURL())
	}
}

func TestNewPostgresConfig_WithoutEnvVar(t *testing.T) {
	// Убираем переменную окружения, если она установлена
	os.Unsetenv("DATABASE_URL")

	// Создаем новый объект конфигурации
	postgresConfig := NewPostgresConfig()

	// Проверяем, что возвращается URL по умолчанию
	expectedURL := defaultDataBaseURL
	if postgresConfig.GetURL() != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, postgresConfig.GetURL())
	}
}
