package usecase

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

var (
	ErrPostDoesNotBelongToUser = errors.New("post does not belong to user")
	ErrPostNotFound            = errors.New("post not found")
	ErrUploadFile              = errors.New("upload file error")
)

type PostRepository interface {
	AddPost(ctx context.Context, post models.Post) error
	UpdatePostText(ctx context.Context, postId uuid.UUID, text string) error
	UpdatePostFiles(ctx context.Context, postId uuid.UUID, fileURLs []string) error
	DeletePost(ctx context.Context, postId uuid.UUID) error
	BelongsTo(ctx context.Context, userId uuid.UUID, postId uuid.UUID) (bool, error)
	GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	GetPostFiles(ctx context.Context, postId uuid.UUID) ([]string, error)
}

type FileRepository interface {
	UploadFile(ctx context.Context, file *models.File) (string, error)
	UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error)
	GetFileURL(ctx context.Context, filename string) (string, error)
	DeleteFile(ctx context.Context, filename string) error
	GetUserAvatar(ctx context.Context, userId uuid.UUID) (string, error)
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
		return fmt.Errorf("p.postRepo.AddPost: %w", err)
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostService) DeletePost(ctx context.Context, user models.User, postId uuid.UUID) error {
	belongsTo, err := p.postRepo.BelongsTo(ctx, user.Id, postId)
	if err != nil {
		return ErrPostNotFound
	}
	if !belongsTo {
		return ErrPostDoesNotBelongToUser
	}

	// retrieve post files
	postFiles, err := p.postRepo.GetPostFiles(ctx, postId)
	if err != nil {
		return fmt.Errorf("p.postRepo.GetPostFiles: %w", err)
	}

	err = p.postRepo.DeletePost(ctx, postId)
	if err != nil {
		return fmt.Errorf("p.postRepo.DeletePost: %w", err)
	}

	// delete post files
	for _, pic := range postFiles {
		err = p.fileRepo.DeleteFile(ctx, path.Base(pic))
		if err != nil {
			return fmt.Errorf("p.fileRepo.DeleteFile: %w", err)
		}
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

func (p *PostService) UpdatePost(ctx context.Context, postUpdate models.PostUpdate, userId uuid.UUID) error {
	// check if user owns the post
	belongsTo, err := p.postRepo.BelongsTo(ctx, userId, postUpdate.Id)
	if err != nil {
		return fmt.Errorf("p.postRepo.BelongsTo: %w", err)
	}

	if !belongsTo {
		return ErrPostDoesNotBelongToUser
	}

	// retrieve old post photos
	oldPics, err := p.postRepo.GetPostFiles(ctx, postUpdate.Id)
	if err != nil {
		return fmt.Errorf("p.postRepo.GetPostFiles: %w", err)
	}

	// Upload files to storage
	var fileURLs []string
	if len(postUpdate.Files) != 0 {
		fileURLs, err = p.fileRepo.UploadManyFiles(ctx, postUpdate.Files)
		if err != nil {
			return fmt.Errorf("p.fileRepo.UploadManyFiles: %w", err)
		}
	}

	err = p.postRepo.UpdatePostText(ctx, postUpdate.Id, postUpdate.Desc)
	if err != nil {
		return fmt.Errorf("p.postRepo.UpdatePostText: %w", err)
	}

	err = p.postRepo.UpdatePostFiles(ctx, postUpdate.Id, fileURLs)
	if err != nil {
		return fmt.Errorf("p.postRepo.UpdatePostFiles: %w", err)
	}

	// delete old photos
	for _, pic := range oldPics {
		err = p.fileRepo.DeleteFile(ctx, path.Base(pic))
		if err != nil {
			return fmt.Errorf("p.fileRepo.DeleteFile: %w", err)
		}
	}

	return nil
}
