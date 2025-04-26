package server_config

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
)

func TestLoadConfig_Success(t *testing.T) {
	// Создаем временный конфигурационный файл для теста
	cfg := &ServerConfig{
		Addr:         "localhost:8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	file, err := os.CreateTemp("", "config_*.toml")
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
	loadedCfg, err := loadConfig(file.Name())
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	// Проверяем, что данные совпадают
	if loadedCfg.Addr != cfg.Addr {
		t.Errorf("expected Addr %v, got %v", cfg.Addr, loadedCfg.Addr)
	}
	if loadedCfg.ReadTimeout != cfg.ReadTimeout {
		t.Errorf("expected ReadTimeout %v, got %v", cfg.ReadTimeout, loadedCfg.ReadTimeout)
	}
	if loadedCfg.WriteTimeout != cfg.WriteTimeout {
		t.Errorf("expected WriteTimeout %v, got %v", cfg.WriteTimeout, loadedCfg.WriteTimeout)
	}
}
