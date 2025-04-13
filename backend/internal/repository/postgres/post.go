package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config/postgres"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/pkg/logger"
)

const getPhotosQuery = `
	select file_url
	from post_file
	where post_id = $1
`

const getOlderPostsLimitQuery = `
	select id, creator_id, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost
	from post 
	where created_at < $1 
	order by created_at desc
	limit $2
`

const insertPostQuery = `
	insert into post (id, creator_id, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

const insertPhotoQuery = `
	insert into post_file (post_id, file_url)
	values ($1, $2)
`

type PostgresPostRepository struct {
	connPool *pgxpool.Pool
}

func NewPostgresPostRepository() *PostgresPostRepository {
	connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &PostgresPostRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (p *PostgresPostRepository) Close() {
	p.connPool.Close()
}

// AddPost adds post to the repository.
func (p *PostgresPostRepository) AddPost(ctx context.Context, post models.Post) error {
	postPostgres := pgmodels.ConvertPostToPostgres(post)
	_, err := p.connPool.Exec(ctx, insertPostQuery,
		postPostgres.Id, postPostgres.CreatorId, postPostgres.Desc,
		postPostgres.CreatedAt, postPostgres.UpdatedAt, postPostgres.LikeCount, postPostgres.RepostCount,
		postPostgres.CommentCount, postPostgres.IsRepost)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to save post %v to database: %s", post, err.Error()))
		return fmt.Errorf("unable to save post to database: %w", err)
	}

	for _, picture := range postPostgres.ImagesURLs {
		_, err = p.connPool.Exec(ctx, insertPhotoQuery,
			postPostgres.Id, picture)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to save post pictures %v for post %v to database: %s",
				postPostgres.ImagesURLs, post, err.Error()))
			return fmt.Errorf("unable to save post pictures to database: %w", err)
		}
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostgresPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
	_, err := p.connPool.Exec(ctx, "delete from post cascade where id = $1", pgtype.UUID{Bytes: postId, Valid: true})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete post %v from database: %s", postId, err.Error()))
		return fmt.Errorf("unable to delete post from database: %w", err)
	}

	return nil
}

func (p *PostgresPostRepository) BelongsTo(ctx context.Context, userId uuid.UUID, postId uuid.UUID) (bool, error) {
	var id uuid.UUID
	err := p.connPool.QueryRow(ctx, "select creator_id from post where id = $1", pgtype.UUID{Bytes: postId, Valid: true}).Scan(&id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post %v from database: %s", postId, err.Error()))
		return false, fmt.Errorf("unable to get post from database: %w", err)
	}

	return id == userId, nil
}

// GetPostsForUId returns posts for user.
func (p *PostgresPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	rows, err := p.connPool.Query(ctx, getOlderPostsLimitQuery, timestamp, numPosts)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get posts from database for user %v, numPosts %v, timestamp %v: %s",
			uid, numPosts, timestamp, err.Error()))
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() {
		var postPostgres pgmodels.PostPostgres
		err = rows.Scan(
			&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.Desc,
			&postPostgres.CreatedAt, &postPostgres.UpdatedAt, &postPostgres.LikeCount,
			&postPostgres.RepostCount, &postPostgres.CommentCount, &postPostgres.IsRepost)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := p.connPool.Query(ctx, getPhotosQuery, postPostgres.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var pic pgtype.Text
			err = pics.Scan(&pic)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", pic, err.Error()))
				return nil, fmt.Errorf("unable to get posts from database: %w", err)
			}

			postPostgres.ImagesURLs = append(postPostgres.ImagesURLs, pic)
		}
		pics.Close()
		result = append(result, postPostgres.ToPost())
	}

	return result, nil
}

func (p *PostgresPostRepository) UpdatePostText(ctx context.Context, postId uuid.UUID, text string) error {
	_, err := p.connPool.Exec(ctx, "update post set text = $1, updated_at = $2 where id = $3", text, time.Now(), postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to update post %v in database: %s", postId, err.Error()))
		return fmt.Errorf("unable to update post in database: %w", err)
	}

	return nil
}

func (p *PostgresPostRepository) UpdatePostFiles(ctx context.Context, postId uuid.UUID, fileURLs []string) error {
	_, err := p.connPool.Exec(ctx, "delete from post_file where post_id = $1", postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete post pictures %v from database: %s", postId, err.Error()))
		return fmt.Errorf("unable to delete post pictures from database: %w", err)
	}

	for _, fileURL := range fileURLs {
		_, err = p.connPool.Exec(ctx, "insert into post_file (post_id, file_url) values ($1, $2)", postId, fileURL)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to insert post picture %v into database: %s", fileURL, err.Error()))
			return fmt.Errorf("unable to insert post picture into database: %w", err)
		}
	}

	return nil
}

func (p *PostgresPostRepository) GetPostFiles(ctx context.Context, postId uuid.UUID) ([]string, error) {
	rows, err := p.connPool.Query(ctx, getPhotosQuery, postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postId, err.Error()))
		return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var pic pgtype.Text
		err = rows.Scan(&pic)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", pic, err.Error()))
			return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
		}

		result = append(result, pic.String)
	}

	return result, nil
}
