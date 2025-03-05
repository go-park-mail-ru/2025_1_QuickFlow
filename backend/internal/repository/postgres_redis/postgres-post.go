package postgres_redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config"
	"quickflow/internal/models"
)

type PostgresPostRepository struct {
}

func NewPostgresPostRepository() *PostgresPostRepository {
	return &PostgresPostRepository{}
}

func (p *PostgresPostRepository) AddPost(ctx context.Context, post models.Post) error {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	// TODO: предусмотреть другой механизм выдачи id
	post.Id = uuid.New()

	_, err = conn.Exec(ctx, "insert into posts values ($1, $2, $3, $4, $5, $6, $7)",
		post.Id, post.CreatorId, post.Desc, post.CreatedAt, post.LikeCount, post.RepostCount, post.CommentCount)
	if err != nil {
		return fmt.Errorf("unable to save user to database: %w", err)
	}

	for _, picture := range post.Pics {
		_, err = conn.Exec(ctx, "insert into post_photos (post_id, photo_path) values ($1, $2)",
			post.Id, picture)
		if err != nil {
			return fmt.Errorf("unable to save user to database: %w", err)
		}
	}

	return nil
}

func (p *PostgresPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
	panic("TODO")
}

func (p *PostgresPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	dbpool, err := pgxpool.New(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	rows, err := dbpool.Query(ctx, "select * from posts where created_at < $1", timestamp)
	if err != nil {
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() && len(result) <= numPosts {
		var post models.Post
		err = rows.Scan(&post.Id, &post.CreatorId, &post.Desc, &post.CreatedAt, &post.LikeCount, &post.RepostCount, &post.CommentCount)
		if err != nil {
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := dbpool.Query(ctx, "select photo_path from post_photos where post_id = $1", post.Id)
		if err != nil {
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var pic string
			err = pics.Scan(&pic)
			if err != nil {
				return nil, fmt.Errorf("unable to get posts from database: %w", err)
			}

			post.Pics = append(post.Pics, pic)
		}
		pics.Close()
		result = append(result, post)
	}

	return result, nil
}
