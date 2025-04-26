package server_config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

const defaultConfigPath = "../deploy/config/feeder/config.toml"

type ServerConfig struct {
	Addr         string        `toml:"addr"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
}

// loadConfig loads config from file.
func loadConfig(configPath string) (*ServerConfig, error) {
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	var cfg ServerConfig
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}
	return &cfg, nil
}

func Parse(configPath string) (*ServerConfig, error) {
	// Loading config
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("internal.Run: %w", err)
	}

	return cfg, nil
}
