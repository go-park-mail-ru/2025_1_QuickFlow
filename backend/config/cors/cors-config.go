package cors_config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type CORSConfig struct {
	AllowedOrigins     []string `toml:"allowed_origins"`
	AllowedMethods     []string `toml:"allowed_methods"`
	AllowedHeaders     []string `toml:"allowed_headers"`
	ExposedHeaders     []string `toml:"exposed_headers"`
	AllowCredentials   bool     `toml:"allow_credentials"`
	OptionsPassthrough bool     `toml:"options_passthrough"`
	Debug              bool     `toml:"debug"`
}

// loadCORSConfig loads config from file.
func loadCORSConfig(configPath string) (*CORSConfig, error) {
	if configPath == "" {
		configPath = "../deploy/config/cors/config.toml"
	}

	var cfg CORSConfig
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}
	return &cfg, nil
}

func ParseCORS(configPath string) (*CORSConfig, error) {
	// Loading config
	cfg, err := loadCORSConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("internal.Run: %w", err)
	}

	return cfg, nil
}
