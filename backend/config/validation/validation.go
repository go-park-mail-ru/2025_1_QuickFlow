package validation_config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
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

	sizeBytes, err := ParseSize(raw.MaxPictureSize)
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

func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "B"):
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	default:
		// по умолчанию — в байтах
	}

	num, err := strconv.ParseFloat(strings.TrimSpace(sizeStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}

	return int64(num * float64(multiplier)), nil
}
