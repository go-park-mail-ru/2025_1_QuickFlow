package minio

import (
	"context"
	"fmt"
	"log"
	thread_safe_map "quickflow/pkg/thread-safe-map"
	"sync"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	minioconfig "quickflow/config/minio"
	"quickflow/internal/models"
)

type MinioRepository struct {
	client *minio.Client
	cfg    *minioconfig.MinioConfig
}

func NewMinioRepository() *MinioRepository {
	cfg := minioconfig.NewMinioConfig()
	client, err := minio.New(cfg.MinioInternalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioRootUser, cfg.MinioRootPassword, ""),
		Secure: cfg.MinioUseSSL,
	})

	if err != nil {
		log.Fatalf("could not create minio client: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.PostsBucketName)
	if err != nil {
		log.Fatalf("could not check if bucket exists: %v", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), cfg.PostsBucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("could not create bucket: %v", err)
		}
	}

	return &MinioRepository{client: client, cfg: cfg}
}

// UploadFile uploads file to MinIO and returns a public URL.
func (m *MinioRepository) UploadFile(ctx context.Context, file models.File) (string, error) {
	_, err := m.client.PutObject(ctx, m.cfg.PostsBucketName, file.Name, file.Reader, file.Size, minio.PutObjectOptions{
		ContentType: file.MimeType,
	})
	if err != nil {
		return "", fmt.Errorf("could not upload file: %v", err)
	}
    
	publicURL := fmt.Sprintf("%s://%s/%s/%s", m.cfg.Scheme, m.cfg.MinioPublicEndpoint, m.cfg.PostsBucketName, file.Name)
	return publicURL, nil
}

// UploadManyFiles uploads multiple files and returns a map of public URLs.
func (m *MinioRepository) UploadManyFiles(ctx context.Context, files []models.File) (map[uuid.UUID]string, error) {
	urls := thread_safe_map.NewThreadSafeMap[uuid.UUID, string]()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for _, file := range files {
		uuID := uuid.New()
		fileName := uuID.String() + file.Ext
		wg.Add(1)

		go func(uuID uuid.UUID, filename string, file models.File) {
			defer wg.Done()
			log.Printf("Adding picture %v of size %d\n", fileName, file.Size)
			_, err := m.client.PutObject(ctx, m.cfg.PostsBucketName, fileName, file.Reader, file.Size, minio.PutObjectOptions{
				ContentType: file.MimeType,
			})
			if err != nil {
				log.Printf("could not upload file: %v, err: %v", file.Name, err)
				cancel()
				return
			}

			publicURL := fmt.Sprintf("%s://%s/%s/%s", m.cfg.Scheme, m.cfg.MinioPublicEndpoint, m.cfg.PostsBucketName, fileName)
			urls.Set(uuID, publicURL)
		}(uuID, fileName, file)
	}

	wg.Wait()
	return urls.GetMapCopy(), nil
}

// GetFileURL returns a public URL for the file.
func (m *MinioRepository) GetFileURL(_ context.Context, fileName string) (string, error) {
	return fmt.Sprintf("%s://%s/%s/%s", m.cfg.Scheme, m.cfg.MinioPublicEndpoint, m.cfg.PostsBucketName, fileName), nil
}

// DeleteFile deletes a file from MinIO.
func (m *MinioRepository) DeleteFile(ctx context.Context, fileName string) error {
	err := m.client.RemoveObject(ctx, m.cfg.PostsBucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("could not delete file: %v", err)
	}
	return nil
}
