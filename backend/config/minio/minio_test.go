package minio_cofig

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestLoadMinioConfig_Success(t *testing.T) {
	cfg := &loadableConfig{
		MinioUseSSL:            false,
		PresignedURLExpiration: 24 * time.Hour,
	}

	// Создаём временный конфигурационный файл для теста
	file, err := os.CreateTemp("", "Minio_config_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Сохраняем конфигурацию в файл
	encoder := toml.NewEncoder(file)
	if err = encoder.Encode(cfg); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Загружаем конфигурацию из файла
	loadedCfg, err := loadMinioConfig(file.Name())
	if err != nil {
		t.Fatalf("loadMinioConfig() failed: %v", err)
	}

	assert.Equal(t, cfg.MinioUseSSL, loadedCfg.MinioUseSSL)
	assert.Equal(t, cfg.PresignedURLExpiration, loadedCfg.PresignedURLExpiration)
}

func TestLoadMinioConfig_FileNotFound(t *testing.T) {
	// Проверим случай, когда файл не существует
	_, err := loadMinioConfig("non_existent_file.toml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
