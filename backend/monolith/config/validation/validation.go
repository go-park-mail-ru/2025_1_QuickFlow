package validation_config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const defaultConfigPath = "../deploy/config/validation/config.toml"

type ValidationConfig struct {
	MaxPostPicturesCount    int      `toml:"max_post_pictures_count"`
	MaxMessagePicturesCount int      `toml:"max_message_pictures_count"`
	AllowedImgExt           []string `toml:"allowed_img_ext"`
	MaxPostPicturesSize     string   `toml:"max_post_pictures_size"`
	MaxMessagePicturesSize  string   `toml:"max_message_pictures_size"`
	MaxPostTextLength       int      `toml:"max_post_text_length"`
	MaxMessageTextLength    int      `toml:"max_message_text_length"`
}

func NewValidationConfig(configPath string) (*ValidationConfig, error) {
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	var cfg ValidationConfig
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to parse validation config from file %v: %w", configPath, err)
	}
	return &cfg, nil
}
