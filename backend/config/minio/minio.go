package minio_config

import (
	getenv "quickflow/utils/get-env"
	"time"
)

type MinioConfig struct {
	Port                   string // Порт, на котором запускается сервер
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

func NewMinioConfig() *MinioConfig {
	return &MinioConfig{
		Port:                   getenv.GetEnv("MINIO_PORT", "8080"),
		MinioInternalEndpoint:  getenv.GetEnv("MINIO_INTERNAL_ENDPOINT", "minio:9000"),
		MinioPublicEndpoint:    getenv.GetEnv("MINIO_PUBLIC_ENDPOINT", "quickflowapp.ru/minio"),
		PostsBucketName:        getenv.GetEnv("MINIO_POSTS_BUCKET_NAME", "posts"),
		ProfileBucketName:      getenv.GetEnv("MINIO_PROFILE_BUCKET_NAME", "profiles"),
		AttachmentsBucketName:  getenv.GetEnv("MINIO_ATTACHMENTS_BUCKET_NAME", "attachments"),
		MinioRootUser:          getenv.GetEnv("MINIO_ROOT_USER", "admin"),
		MinioRootPassword:      getenv.GetEnv("MINIO_ROOT_PASSWORD", "adminpassword"),
		Scheme:                 getenv.GetEnv("MINIO_SCHEME", "https"),
		MinioUseSSL:            getenv.GetEnvAsBool("MINIO_USE_SSL", false),
		PresignedURLExpiration: 24 * time.Hour,
	}
}

// https://habr.com/ru/articles/818853/
