package validation_config

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"quickflow/file_service/config/validation"
)

const defaultConfigPath = "../deploy/config/validation/config.toml"

type validationConfigRaw struct {
	MaxPostPicturesCount    int      `toml:"max_post_pictures_count"`
	MaxMessagePicturesCount int      `toml:"max_message_pictures_count"`
	AllowedImgExt           []string `toml:"allowed_img_ext"`
	MaxPostPicturesSize     string   `toml:"max_post_pictures_size"`
	MaxMessagePicturesSize  string   `toml:"max_message_pictures_size"`
	MaxPostTextLength       int      `toml:"max_post_text_length"`
	MaxMessageTextLength    int      `toml:"max_message_text_length"`
	MaxPictureCount         int      `toml:"max_picture_count"`
	MaxPictureSize          string   `toml:"max_picture_size"`
}

type ValidationConfig struct {
	MaxPostPicturesCount    int
	MaxMessagePicturesCount int
	AllowedImgExt           []string
	MaxPostPicturesSize     string
	MaxMessagePicturesSize  string
	MaxPostTextLength       int
	MaxMessageTextLength    int
	MaxPictureCount         int
	MaxPictureSize          int64
}

func NewValidationConfig(configPath string) (*ValidationConfig, error) {
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	var raw validationConfigRaw
	_, err := toml.DecodeFile(configPath, &raw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse validation config from file %v: %w", configPath, err)
	}

	sizeBytes, err := validation.ParseSize(raw.MaxPictureSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_picture_size format: %w", err)
	}

	cfg := &ValidationConfig{
		MaxPostPicturesCount:    raw.MaxPostPicturesCount,
		MaxMessagePicturesCount: raw.MaxMessagePicturesCount,
		AllowedImgExt:           raw.AllowedImgExt,
		MaxPostPicturesSize:     raw.MaxPostPicturesSize,
		MaxMessagePicturesSize:  raw.MaxMessagePicturesSize,
		MaxPostTextLength:       raw.MaxPostTextLength,
		MaxMessageTextLength:    raw.MaxMessageTextLength,
		MaxPictureCount:         raw.MaxPictureCount,
		MaxPictureSize:          sizeBytes,
	}

	return cfg, nil
}
