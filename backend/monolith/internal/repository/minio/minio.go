package minio

import (
	"context"
	"fmt"
	minioconfig "quickflow/monolith/config/minio"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/pkg/logger"
	threadsafeslice "quickflow/monolith/pkg/thread-safe-slice"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

type MinioRepository struct {
	client                *minio.Client
	PostsBucketName       string
	AttachmentsBucketName string
	ProfileBucketName     string
	PublicUrlRoot         string
}

func NewMinioRepository(cfg *minioconfig.MinioConfig) (*MinioRepository, error) {
	client, err := minio.New(cfg.MinioInternalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioRootUser, cfg.MinioRootPassword, ""),
		Secure: cfg.MinioUseSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("could not create minio client: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.PostsBucketName)
	if err != nil {
		return nil, fmt.Errorf("could not check if bucket exists: %v", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), cfg.PostsBucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("could not create bucket: %v", err)
		}
	}

	return &MinioRepository{
		client:                client,
		PostsBucketName:       cfg.PostsBucketName,
		AttachmentsBucketName: cfg.AttachmentsBucketName,
		ProfileBucketName:     cfg.ProfileBucketName,
		PublicUrlRoot:         fmt.Sprintf("%s://%s", cfg.Scheme, cfg.MinioPublicEndpoint),
	}, nil
}

// UploadFile uploads file to MinIO and returns a public URL.
func (m *MinioRepository) UploadFile(ctx context.Context, file *models.File) (string, error) {
	uuID := uuid.New()
	fileName := uuID.String() + file.Ext

	_, err := m.client.PutObject(ctx, m.PostsBucketName, fileName, file.Reader, file.Size, minio.PutObjectOptions{
		ContentType: file.MimeType,
	})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("could not upload file %v: %v", file.Name, err))
		return "", fmt.Errorf("could not upload file: %v", err)
	}

	publicURL := fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, m.PostsBucketName, fileName)
	logger.Info(ctx, fmt.Sprintf("File successfully loaded: %v, url: %v", file.Name, publicURL))
	return publicURL, nil
}

// UploadManyFiles uploads multiple files and returns a map of public URLs.
func (m *MinioRepository) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	urls := threadsafeslice.NewThreadSafeSliceN[string](len(files))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	for i, file := range files {
		i := i
		file := file // https://golang.org/doc/faq#closures_and_goroutines
		uuID := uuid.New()
		fileName := uuID.String() + file.Ext

		wg.Go(func() error {
			_, err := m.client.PutObject(ctx, m.PostsBucketName, fileName, file.Reader, file.Size, minio.PutObjectOptions{
				ContentType: file.MimeType,
			})
			if err != nil {
				return fmt.Errorf("could not upload file: %v, err: %v", file.Name, err)
			}

			publicURL := fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, m.PostsBucketName, fileName)
			err = urls.SetByIdx(i, publicURL)
			if err != nil {
				return fmt.Errorf("could not upload file: %v, err: %v", file.Name, err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}
	return urls.GetSliceCopy(), nil
}

// GetFileURL returns a public URL for the file.
func (m *MinioRepository) GetFileURL(_ context.Context, fileName string) (string, error) {
	return fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, m.PostsBucketName, fileName), nil
}

// DeleteFile deletes a file from MinIO.
func (m *MinioRepository) DeleteFile(ctx context.Context, fileName string) error {
	err := m.client.RemoveObject(ctx, m.PostsBucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("could not delete file: %v", err)
	}
	return nil
}
