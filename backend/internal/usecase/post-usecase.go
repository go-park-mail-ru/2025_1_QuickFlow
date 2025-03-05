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

type PostService struct {
	postRepo PostRepository
}

func NewPostService(postRepo PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}

// AddPost adds post to the repository.
func (p *PostService) AddPost(ctx context.Context, post models.Post) error {
	err := p.postRepo.AddPost(ctx, post)
	if err != nil {
		return fmt.Errorf("p.repo.AddPost: %w", err)
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostService) DeletePost(ctx context.Context, postId uuid.UUID) error {
	err := p.postRepo.DeletePost(ctx, postId)
	if err != nil {
		return fmt.Errorf("p.repo.AddPost: %w", err)
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
