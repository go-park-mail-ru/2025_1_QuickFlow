package usecase

import (
	"context"
	"errors"
	"fmt"
	"path"
	"slices"
	"time"

	"github.com/google/uuid"

	"quickflow/gateway/utils/validation"
	post_errors "quickflow/post_service/internal/errors"
	"quickflow/shared/models"
)

type CommentRepository interface {
	AddComment(ctx context.Context, comment models.Comment) error
	GetCommentFiles(ctx context.Context, commentId uuid.UUID) ([]string, error)
	DeleteComment(ctx context.Context, commentId uuid.UUID) error
	GetCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error)
	GetComment(ctx context.Context, commentId uuid.UUID) (models.Comment, error)
	CheckIfCommentLiked(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (bool, error)
	LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	UpdateComment(ctx context.Context, commentUpdate models.CommentUpdate) error
	GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error)
}

type CommentUseCase struct {
	commentRepo CommentRepository
	fileService FileService
	validator   PostValidator
}

func NewCommentUseCase(commentRepo CommentRepository, fileService FileService, validator PostValidator) *CommentUseCase {
	return &CommentUseCase{commentRepo: commentRepo,
		fileService: fileService,
		validator:   validator,
	}
}

// AddComment adds Comment to the repository.
func (c *CommentUseCase) AddComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	comment.Id = uuid.New()

	var err error
	err = c.commentRepo.AddComment(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.AddComment: %w", err)
	}

	newComment, err := c.commentRepo.GetComment(ctx, comment.Id)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.GetComment: %w", err)
	}
	return &newComment, nil
}

// DeleteComment removes Comment from the repository.
func (c *CommentUseCase) DeleteComment(ctx context.Context, userId uuid.UUID, commentId uuid.UUID) error {
	//belongsTo, err := p.fileService.BelongsTo(ctx, userId, commentId)
	//if err != nil {
	//	return post_errors.ErrCommentNotFound
	//}

	// TODO user_service
	//if !belongsTo && user.Username != "Nikita" && user.Username != "rvasutenko" {
	//	return post_errors.ErrCommentDoesNotBelongToUser
	//}

	// retrieve Comment files
	postFiles, err := c.commentRepo.GetCommentFiles(ctx, commentId)
	if err != nil {
		return fmt.Errorf("p.fileService.GetCommentFiles: %w", err)
	}

	err = c.commentRepo.DeleteComment(ctx, commentId)
	if err != nil {
		return fmt.Errorf("p.fileService.DeleteComment: %w", err)
	}

	// delete Comment files
	for _, pic := range postFiles {
		err = c.fileService.DeleteFile(ctx, path.Base(pic))
		if err != nil {
			return fmt.Errorf("p.fileService.DeleteFile: %w", err)
		}
	}

	return nil
}

func (c *CommentUseCase) FetchCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error) {
	// validate params
	err := c.validator.ValidateFeedParams(numComments, timestamp)
	if errors.Is(err, validation.ErrInvalidNumPosts) {
		return nil, post_errors.ErrInvalidNumComments
	} else if errors.Is(err, validation.ErrInvalidTimestamp) {
		return nil, post_errors.ErrInvalidTimestamp
	} else if err != nil {
		return nil, fmt.Errorf("validation.ValidateFeedParams: %w", err)
	}

	// fetch posts
	posts, err := c.commentRepo.GetCommentsForPost(ctx, postId, numComments, timestamp)
	if err != nil {
		return nil, fmt.Errorf("p.repo.GetCommentsForUId: %w", err)
	}

	return posts, nil
}

func (c *CommentUseCase) UpdateComment(ctx context.Context, commentUpdate models.CommentUpdate, userId uuid.UUID) (*models.Comment, error) {
	//// check if user owns the Comment
	//belongsTo, err := p.fileService.BelongsTo(ctx, userId, commentUpdate.Id)
	//if err != nil {
	//	return nil, fmt.Errorf("p.fileService.BelongsTo: %w", err)
	//}
	//
	//

	//if !belongsTo {
	//	return nil, post_errors.ErrCommentDoesNotBelongToUser
	//}

	oldComment, err := c.commentRepo.GetComment(ctx, commentUpdate.Id)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.GetComment: %w", err)
	}

	if oldComment.UserId != userId {
		return nil, post_errors.ErrDoesNotBelongToUser
	}

	// retrieve old Comment photos
	oldPics, err := c.commentRepo.GetCommentFiles(ctx, commentUpdate.Id)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.GetCommentFiles: %w", err)
	}

	// update Comment
	err = c.commentRepo.UpdateComment(ctx, commentUpdate)

	// delete old photos if they are not in the new photos
	var fileURLs []string
	for _, file := range commentUpdate.Files {
		fileURLs = append(fileURLs, file.URL)
	}

	for _, pic := range oldPics {
		if !slices.Contains(fileURLs, pic) {
			err = c.fileService.DeleteFile(ctx, path.Base(pic))
			if err != nil {
				return nil, fmt.Errorf("p.fileService.DeleteFile: %w", err)
			}
		}
	}

	Comment, err := c.commentRepo.GetComment(ctx, commentUpdate.Id)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.GetComment: %w", err)
	}

	return &Comment, nil
}

func (c *CommentUseCase) LikeComment(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	if userId == uuid.Nil || postId == uuid.Nil {
		return errors.New("userId or postId is empty")
	}

	liked, err := c.commentRepo.CheckIfCommentLiked(ctx, postId, userId)
	if err != nil {
		return fmt.Errorf("p.fileService.CheckIfCommentLiked: %w", err)
	}

	if liked {
		return nil // идемпотентность лайка
	}

	err = c.commentRepo.LikeComment(ctx, postId, userId)
	if err != nil {
		return fmt.Errorf("p.fileService.LikeComment: %w", err)
	}

	return nil
}

func (c *CommentUseCase) UnlikeComment(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	if userId == uuid.Nil || postId == uuid.Nil {
		return errors.New("userId or postId is empty")
	}

	liked, err := c.commentRepo.CheckIfCommentLiked(ctx, postId, userId)
	if err != nil {
		return fmt.Errorf("p.fileService.CheckIfCommentLiked: %w", err)
	}

	if !liked {
		return nil // идемпотентность дизлайка
	}

	err = c.commentRepo.UnlikeComment(ctx, postId, userId)
	if err != nil {
		return fmt.Errorf("p.fileService.UnlikeComment: %w", err)
	}

	return nil
}

func (c *CommentUseCase) GetComment(ctx context.Context, postId, userId uuid.UUID) (*models.Comment, error) {
	if postId == uuid.Nil {
		return nil, fmt.Errorf("postId is empty")
	}

	Comment, err := c.commentRepo.GetComment(ctx, postId)
	if err != nil {
		if errors.Is(err, post_errors.ErrNotFound) {
			return nil, post_errors.ErrNotFound
		}
		return nil, fmt.Errorf("p.fileService.GetComment: %w", err)
	}
	return &Comment, err
}

func (c *CommentUseCase) GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error) {
	if postId == uuid.Nil {
		return nil, fmt.Errorf("postId is empty")
	}

	Comment, err := c.commentRepo.GetLastPostComment(ctx, postId)
	if err != nil {
		return nil, fmt.Errorf("p.fileService.GetComment: %w", err)
	}
	return Comment, err
}
