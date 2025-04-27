package validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const defaultConfigPath = "../deploy/config/validation/config.toml"

type ValidationConfig struct {
	MaxPicturesCount int
	AllowedFileExt   []string

	MaxPictureSize int64
}

type validationConfigRaw struct {
	MaxPicturesCount int      `toml:"max_picture_count"`
	AllowedFileExt   []string `toml:"allowed_file_ext"`
	MaxPictureSize   string   `toml:"max_picture_size"`
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
		MaxPicturesCount: raw.MaxPicturesCount,
		AllowedFileExt:   raw.AllowedFileExt,
		MaxPictureSize:   sizeBytes,
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
