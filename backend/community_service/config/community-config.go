package config

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"quickflow/file_service/config/validation"
)

const defaultConfigPath = "../deploy/config/community/config.toml"

type communityConfigRaw struct {
	CommunityNameMinLength        int    `toml:"community_name_min_length"`
	CommunityNameMaxLength        int    `toml:"community_name_max_length"`
	CommunityDescriptionMaxLength int    `toml:"community_description_max_length"`
	CommunityAvatarMaxSize        string `toml:"community_avatar_max_size"`
}

type CommunityConfig struct {
	CommunityNameMinLength        int
	CommunityNameMaxLength        int
	CommunityDescriptionMaxLength int
	CommunityAvatarMaxSize        int64
}

func NewCommunityConfig(configPath string) (*CommunityConfig, error) {
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	var raw communityConfigRaw
	_, err := toml.DecodeFile(configPath, &raw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse validation config from file %v: %w", configPath, err)
	}

	sizeBytes, err := validation.ParseSize(raw.CommunityAvatarMaxSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max_picture_size format: %w", err)
	}

	cfg := &CommunityConfig{
		CommunityNameMinLength:        raw.CommunityNameMinLength,
		CommunityNameMaxLength:        raw.CommunityNameMaxLength,
		CommunityDescriptionMaxLength: raw.CommunityDescriptionMaxLength,
		CommunityAvatarMaxSize:        sizeBytes,
	}

	return cfg, nil
}
