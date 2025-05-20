package validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const defaultConfigPath = "../deploy/config/validation/config.toml"

type ValidationConfig struct {
	MaxFileCount int

	MaxPictureSize int64
	MaxVideoSize   int64
	MaxAudioSize   int64
	MaxFileSize    int64

	AllowedVideoExt   []string
	AllowedPictureExt []string
	AllowedFileExt    []string
	AllowedAudioExt   []string
}

type validationConfigRaw struct {
	MaxFileCount int `toml:"max_file_count"`

	MaxPictureSize string `toml:"max_picture_size"`
	MaxVideoSize   string `toml:"max_video_size"`
	MaxAudioSize   string `toml:"max_audio_size"`
	MaxFileSize    string `toml:"max_file_size"`

	AllowedVideoExt   []string `toml:"allowed_video_ext"`
	AllowedPictureExt []string `toml:"allowed_picture_ext"`
	AllowedAudioExt   []string `toml:"allowed_audio_ext"`
	AllowedFileExt    []string `toml:"allowed_file_ext,omitempty"`
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

	picSize, err := ParseSize(raw.MaxPictureSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_picture_size format: %w", err)
	}
	videoSize, err := ParseSize(raw.MaxVideoSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_video_size format: %w", err)
	}
	audioSize, err := ParseSize(raw.MaxAudioSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_audio_size format: %w", err)
	}
	fileSize, err := ParseSize(raw.MaxFileSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_file_size format: %w", err)
	}

	cfg := &ValidationConfig{
		MaxFileCount: raw.MaxFileCount,

		AllowedFileExt:    raw.AllowedFileExt,
		MaxFileSize:       fileSize,
		MaxPictureSize:    picSize,
		AllowedPictureExt: raw.AllowedPictureExt,
		MaxVideoSize:      videoSize,
		AllowedVideoExt:   raw.AllowedVideoExt,
		MaxAudioSize:      audioSize,
		AllowedAudioExt:   raw.AllowedAudioExt,
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
