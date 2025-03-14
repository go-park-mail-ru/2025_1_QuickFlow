package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
)

const getPhotosQuery = `
	select photo_path
	from post_photos
	where post_id = $1
`

const getOlderPostsLimitQuery = `
	select * 
	from posts 
	where created_at < $1 
	order by created_at 
	limit $2
`

const insertPostQuery = `
	insert into posts (id, creator_id, desc, created_at, like_count, repost_count, comment_count)
	values ($1, $2, $3, $4, $5, $6, $7)
`

const insertPhotoQuery = `
	insert into post_photos (post_id, photo_path)
	values ($1, $2)
`

type PostgresPostRepository struct {
}

func NewPostgresPostRepository() *PostgresPostRepository {
	return &PostgresPostRepository{}
}

// AddPost adds post to the repository.
func (p *PostgresPostRepository) AddPost(ctx context.Context, post models.Post) error {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	postPostgres := pgmodels.ConvertPostToPostgres(post)

	_, err = conn.Exec(ctx, insertPostQuery,
		postPostgres.Id, postPostgres.CreatorId, postPostgres.Desc,
		postPostgres.CreatedAt, postPostgres.LikeCount, postPostgres.RepostCount,
		postPostgres.CommentCount)
	if err != nil {
		return fmt.Errorf("unable to save user to database: %w", err)
	}

	for _, picture := range postPostgres.Pics {
		_, err = conn.Exec(ctx, insertPhotoQuery,
			postPostgres.Id, picture)
		if err != nil {
			return fmt.Errorf("unable to save user to database: %w", err)
		}
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostgresPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "delete from posts where id = $1", pgtype.UUID{Bytes: postId, Valid: true})
	if err != nil {
		return fmt.Errorf("unable to delete post from database: %w", err)
	}

	return nil
}

// GetPostsForUId returns posts for user.
func (p *PostgresPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	dbPool, err := pgxpool.New(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbPool.Close()

	rows, err := dbPool.Query(ctx, getOlderPostsLimitQuery, timestamp, numPosts)
	if err != nil {
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() {
		var postPostgres pgmodels.PostPostgres
		err = rows.Scan(
			&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.Desc,
			&postPostgres.CreatedAt, &postPostgres.LikeCount,
			&postPostgres.RepostCount, &postPostgres.CommentCount)
		if err != nil {
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := dbPool.Query(ctx, getPhotosQuery, postPostgres.Id)
		if err != nil {
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var pic pgtype.Text
			err = pics.Scan(&pic)
			if err != nil {
				return nil, fmt.Errorf("unable to get posts from database: %w", err)
			}

			postPostgres.Pics = append(postPostgres.Pics, pic)
		}
		pics.Close()
		result = append(result, postPostgres.ToPost())
	}

	return result, nil
}
