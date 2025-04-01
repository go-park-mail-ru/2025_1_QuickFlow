package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

type PostRepository interface {
	AddPost(ctx context.Context, post models.Post) error
	DeletePost(ctx context.Context, postId uuid.UUID) error
	GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
}

type FileRepository interface {
	UploadFile(ctx context.Context, file *models.File) (string, error)
	UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error)
	GetFileURL(ctx context.Context, filename string) (string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type PostService struct {
	postRepo PostRepository
	fileRepo FileRepository
}

// NewPostService creates new post service.
func NewPostService(postRepo PostRepository, fileRepo FileRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		fileRepo: fileRepo,
	}
}

// AddPost adds post to the repository.
func (p *PostService) AddPost(ctx context.Context, post models.Post) error {
	post.Id = uuid.New()

	var err error
	// Upload files to storage
	post.ImagesURL, err = p.fileRepo.UploadManyFiles(ctx, post.Images)
	if err != nil {
		return fmt.Errorf("p.fileRepo.UploadManyFiles: %w", err)
	}

	// Update post images with urls
	err = p.postRepo.AddPost(ctx, post)
	if err != nil {
		return fmt.Errorf("p.repo.AddPost: %w", err)
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostService) DeletePost(ctx context.Context, postId uuid.UUID) error {
	err := p.postRepo.DeletePost(ctx, postId)
	if err != nil {
		return fmt.Errorf("p.repo.DeletePost: %w", err)
	}

	return nil
}

// FetchFeed returns feed for user.
func (p *PostService) FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error) {
	posts, err := p.postRepo.GetPostsForUId(ctx, user.Id, numPosts, timestamp)
	if err != nil {
		return []models.Post{}, fmt.Errorf("p.repo.GetPostsForUId: %w", err)
	}

	return posts, nil
}
