package cors_config

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestLoadCORSConfig_Success(t *testing.T) {
	// Создаём временный конфигурационный файл для теста
	cfg := &CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST"},
		AllowedHeaders:     []string{"Content-Type"},
		ExposedHeaders:     []string{"X-My-Header"},
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              false,
	}

	file, err := os.CreateTemp("", "cors_config_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Сохраняем конфигурацию в файл
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Загружаем конфигурацию из файла
	loadedCfg, err := loadCORSConfig(file.Name())
	if err != nil {
		t.Fatalf("loadCORSConfig() failed: %v", err)
	}

	// Проверяем, что данные совпадают
	if len(loadedCfg.AllowedOrigins) != len(cfg.AllowedOrigins) {
		t.Errorf("expected AllowedOrigins %v, got %v", cfg.AllowedOrigins, loadedCfg.AllowedOrigins)
	}
	if len(loadedCfg.AllowedMethods) != len(cfg.AllowedMethods) {
		t.Errorf("expected AllowedMethods %v, got %v", cfg.AllowedMethods, loadedCfg.AllowedMethods)
	}
}

func TestLoadCORSConfig_FileNotFound(t *testing.T) {
	// Проверим случай, когда файл не существует
	_, err := loadCORSConfig("non_existent_file.toml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
