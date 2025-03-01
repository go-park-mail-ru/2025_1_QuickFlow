package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Addr         string        `toml:"addr"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
}

// loadConfig loads config from file.
func loadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "../deploy/config/feeder/config.toml"
	}

	var cfg Config
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}
	return &cfg, nil
}

func Parse() (*Config, error) {
	// Supporting config path via flags
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Loading config
	cfg, err := loadConfig(*configPath)
	if err != nil {
		return nil, fmt.Errorf("internal.Run: %w", err)
	}

	return cfg, nil
}
