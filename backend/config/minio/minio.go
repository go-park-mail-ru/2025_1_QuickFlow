package minio_config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"

	getenv "quickflow/utils/get-env"
)

const (
	defaultMinioInternalEndpoint = "minio:9000"
	defaultMinioPublicEndpoint   = "quickflowapp.ru/minio"
	defaultPostsBucketName       = "posts"
	defaultProfileBucketName     = "profiles"
	defaultAttachmentsBucketName = "attachments"
	defaultMinioRootUser         = "admin"
	defaultMinioRootPassword     = "adminpassword"
	defaultScheme                = "https"
)

type MinioConfig struct {
	MinioInternalEndpoint  string // Адрес конечной точки Minio
	MinioPublicEndpoint    string
	PostsBucketName        string // Название конкретного бакета в Minio
	ProfileBucketName      string
	AttachmentsBucketName  string
	MinioRootUser          string // Имя пользователя для доступа к Minio
	MinioRootPassword      string // Пароль для доступа к Minio
	MinioUseSSL            bool
	PresignedURLExpiration time.Duration
	Scheme                 string
}

type loadableConfig struct {
	PresignedURLExpiration time.Duration `toml:"presigned_url_expiration"`
	MinioUseSSL            bool          `toml:"minio_use_ssl"`
}

// loadMinioConfig loads config from file.
func loadMinioConfig(configPath string) (*MinioConfig, error) {
	if configPath == "" {
		configPath = "../deploy/config/minio/config.toml"
	}

	var cfg loadableConfig
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	return newMinioConfig(cfg), nil
}

func newMinioConfig(config loadableConfig) *MinioConfig {
	return &MinioConfig{
		MinioInternalEndpoint:  getenv.GetEnv("MINIO_INTERNAL_ENDPOINT", defaultMinioInternalEndpoint),
		MinioPublicEndpoint:    getenv.GetEnv("MINIO_PUBLIC_ENDPOINT", defaultMinioPublicEndpoint),
		PostsBucketName:        getenv.GetEnv("MINIO_POSTS_BUCKET_NAME", defaultPostsBucketName),
		ProfileBucketName:      getenv.GetEnv("MINIO_PROFILE_BUCKET_NAME", defaultProfileBucketName),
		AttachmentsBucketName:  getenv.GetEnv("MINIO_ATTACHMENTS_BUCKET_NAME", defaultAttachmentsBucketName),
		MinioRootUser:          getenv.GetEnv("MINIO_ROOT_USER", defaultMinioRootUser),
		MinioRootPassword:      getenv.GetEnv("MINIO_ROOT_PASSWORD", defaultMinioRootPassword),
		Scheme:                 getenv.GetEnv("MINIO_SCHEME", defaultScheme),
		MinioUseSSL:            config.MinioUseSSL,
		PresignedURLExpiration: config.PresignedURLExpiration,
	}
}

func ParseMinio(configPath string) (*MinioConfig, error) {
	// Loading config
	cfg, err := loadMinioConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("internal.Run: %w", err)
	}

	return cfg, nil
}
