package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	post_errors "quickflow/post_service/internal/errors"
	pgmodels "quickflow/post_service/internal/repository/postgres-models"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

const getPostsQuery = `
	select p.id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost
	from post p
	where p.id = $1
`
const getPhotosQuery = `
	select pf.file_url, pf.file_type, f.filename
	from post_file pf
	join files f 
	on pf.file_url = f.file_url
	where post_id = $1
	order by added_at;
`

const getRecommendationsForUserOlder = `
	select id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost
	from post 
	where created_at < $1 
	order by created_at desc
	limit $2;
`

const getUserPostsOlder = `
	select id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost
	from post
	where creator_id = $1 and created_at < $2
	order by created_at desc
	limit $3;
`

const getPostsForUserOlder = `
	with followed_by_user as (
		select user1_id as id
		from friendship
		where user2_id = $1 and (status = $4 or status = $5) -- friends or followed_by
        union
		select user2_id as id
		from friendship
		where user1_id = $1 and (status = $4 or status = $6) -- friends or following
		union
		select $1 as id
		union
		select community_id as id
		from community_user
		where user_id = $1
	)
	select p.id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost
	from post p
	join followed_by_user fbu on p.creator_id = fbu.id
	where created_at < $2
	order by created_at desc
	limit $3;
`

const insertPostQuery = `
	insert into post (id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const insertPhotoQuery = `
	insert into post_file (post_id, file_url, file_type)
	values ($1, $2, $3)
`

const checkIfPostLikedRequest = `
	select 1
	from like_post
	where post_id = $1 and user_id = $2;
`

const likePostRequest = `
	insert into like_post (user_id, post_id)
	values ($1, $2);
`

const unlikePostRequest = `
	delete from like_post
	where post_id = $1 and user_id = $2;
`

type PostgresPostRepository struct {
	connPool *sql.DB
}

func NewPostgresPostRepository(connPool *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{
		connPool: connPool,
	}
}

// Close закрывает пул соединений
func (p *PostgresPostRepository) Close() {
	p.connPool.Close()
}

// AddPost adds post to the repository.
func (p *PostgresPostRepository) AddPost(ctx context.Context, post models.Post) error {
	postPostgres := pgmodels.ConvertPostToPostgres(post)
	_, err := p.connPool.ExecContext(ctx, insertPostQuery,
		postPostgres.Id, postPostgres.CreatorId, postPostgres.CreatorType, postPostgres.Desc,
		postPostgres.CreatedAt, postPostgres.UpdatedAt, postPostgres.LikeCount, postPostgres.RepostCount,
		postPostgres.CommentCount, postPostgres.IsRepost)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to save post %v to database: %s", post, err.Error()))
		return fmt.Errorf("unable to save post to database: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("Post %v saved to database", postPostgres.Id))

	for _, picture := range postPostgres.Files {
		_, err = p.connPool.ExecContext(ctx, insertPhotoQuery,
			postPostgres.Id, picture.URL, picture.DisplayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to save post pictures %v for post %v to database: %s",
				postPostgres.Files, post, err.Error()))
			return fmt.Errorf("unable to save post pictures to database: %w", err)
		}
	}

	return nil
}

// DeletePost removes post from the repository.
func (p *PostgresPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
	_, err := p.connPool.ExecContext(ctx, "delete from post cascade where id = $1", pgtype.UUID{Bytes: postId, Valid: true})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete post %v from database: %s", postId, err.Error()))
		return fmt.Errorf("unable to delete post from database: %w", err)
	}

	return nil
}

func (p *PostgresPostRepository) BelongsTo(ctx context.Context, userId uuid.UUID, postId uuid.UUID) (bool, error) {
	var id uuid.UUID
	err := p.connPool.QueryRowContext(ctx, "select creator_id from post where id = $1", pgtype.UUID{Bytes: postId, Valid: true}).Scan(&id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post %v from database: %s", postId, err.Error()))
		return false, fmt.Errorf("unable to get post from database: %w", err)
	}

	return id == userId, nil
}

func (p *PostgresPostRepository) GetPost(ctx context.Context, postId uuid.UUID) (models.Post, error) {
	row := p.connPool.QueryRowContext(ctx, getPostsQuery, postId)
	var postPostgres pgmodels.PostPostgres
	err := row.Scan(
		&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.CreatorType, &postPostgres.Desc,
		&postPostgres.CreatedAt, &postPostgres.UpdatedAt, &postPostgres.LikeCount,
		&postPostgres.RepostCount, &postPostgres.CommentCount, &postPostgres.IsRepost)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post %v from database: %s", postId, err.Error()))
		return models.Post{}, fmt.Errorf("unable to get post from database: %w", err)
	}

	pics, err := p.connPool.QueryContext(ctx, getPhotosQuery, postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postId, err.Error()))
		return models.Post{}, fmt.Errorf("unable to get post pictures from database: %w", err)
	}

	for pics.Next() {
		var postgresFile pgmodels.PostgresFile
		err = pics.Scan(&postgresFile.URL, &postgresFile.DisplayType, &postgresFile.Name)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", postgresFile.Name, err.Error()))
			return models.Post{}, fmt.Errorf("unable to get post pictures from database: %w", err)
		}

		postPostgres.Files = append(postPostgres.Files, postgresFile)
	}
	pics.Close()

	return postPostgres.ToPost(), nil
}

func (p *PostgresPostRepository) GetUserPosts(ctx context.Context, id uuid.UUID, requesterId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	rows, err := p.connPool.QueryContext(ctx, getUserPostsOlder, id, timestamp, numPosts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, post_errors.ErrNotFound
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get posts from database for user %v, numPosts %v, timestamp %v: %s",
			id, numPosts, timestamp, err.Error()))
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() {
		var postPostgres pgmodels.PostPostgres
		err = rows.Scan(
			&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.CreatorType, &postPostgres.Desc,
			&postPostgres.CreatedAt, &postPostgres.UpdatedAt, &postPostgres.LikeCount,
			&postPostgres.RepostCount, &postPostgres.CommentCount, &postPostgres.IsRepost)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := p.connPool.QueryContext(ctx, getPhotosQuery, postPostgres.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var postgresFile pgmodels.PostgresFile
			err = pics.Scan(&postgresFile.URL, &postgresFile.DisplayType, &postgresFile.Name)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", postgresFile.Name, err.Error()))
				return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
			}

			postPostgres.Files = append(postPostgres.Files, postgresFile)
		}
		pics.Close()

		// check if requester liked the post
		liked, err := p.CheckIfPostLiked(ctx, postPostgres.Id.Bytes, requesterId)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to check if post %v is liked by user %v: %s", postPostgres.Id, requesterId, err.Error()))
			return nil, fmt.Errorf("unable to check if post is liked by user: %w", err)
		}
		postPostgres.IsLiked = pgtype.Bool{Bool: liked, Valid: true}

		result = append(result, postPostgres.ToPost())
	}

	return result, nil
}

func (p *PostgresPostRepository) GetRecommendationsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	rows, err := p.connPool.QueryContext(ctx, getRecommendationsForUserOlder, timestamp, numPosts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, post_errors.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get posts from database for user %v, numPosts %v, timestamp %v: %s",
			uid, numPosts, timestamp, err.Error()))
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() {
		var postPostgres pgmodels.PostPostgres
		err = rows.Scan(
			&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.CreatorType, &postPostgres.Desc,
			&postPostgres.CreatedAt, &postPostgres.UpdatedAt, &postPostgres.LikeCount,
			&postPostgres.RepostCount, &postPostgres.CommentCount, &postPostgres.IsRepost)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := p.connPool.QueryContext(ctx, getPhotosQuery, postPostgres.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var postgresFile pgmodels.PostgresFile
			err = pics.Scan(&postgresFile.URL, &postgresFile.DisplayType, &postgresFile.Name)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", postgresFile.Name, err.Error()))
				return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
			}

			postPostgres.Files = append(postPostgres.Files, postgresFile)
		}
		pics.Close()

		// check if requester liked the post
		liked, err := p.CheckIfPostLiked(ctx, postPostgres.Id.Bytes, uid)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to check if post %v is liked by user %v: %s", postPostgres.Id, uid, err.Error()))
			return nil, fmt.Errorf("unable to check if post is liked by user: %w", err)
		}
		postPostgres.IsLiked = pgtype.Bool{Bool: liked, Valid: true}

		result = append(result, postPostgres.ToPost())
	}

	return result, nil
}

func (p *PostgresPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	rows, err := p.connPool.QueryContext(ctx, getPostsForUserOlder, uid, timestamp, numPosts,
		models.RelationFriend, models.RelationFollowedBy, models.RelationFollowing)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, post_errors.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get posts from database for user %v, numPosts %v, timestamp %v: %s",
			uid, numPosts, timestamp, err.Error()))
		return nil, fmt.Errorf("unable to get posts from database: %w", err)
	}
	defer rows.Close()

	var result []models.Post
	for rows.Next() {
		var postPostgres pgmodels.PostPostgres
		err = rows.Scan(
			&postPostgres.Id, &postPostgres.CreatorId, &postPostgres.CreatorType, &postPostgres.Desc,
			&postPostgres.CreatedAt, &postPostgres.UpdatedAt, &postPostgres.LikeCount,
			&postPostgres.RepostCount, &postPostgres.CommentCount, &postPostgres.IsRepost)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		pics, err := p.connPool.QueryContext(ctx, getPhotosQuery, postPostgres.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get posts from database: %w", err)
		}

		for pics.Next() {
			var postgresFile pgmodels.PostgresFile
			err = pics.Scan(&postgresFile.URL, &postgresFile.DisplayType, &postgresFile.Name)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", postgresFile.Name, err.Error()))
				return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
			}

			postPostgres.Files = append(postPostgres.Files, postgresFile)
		}
		pics.Close()

		// check if requester liked the post
		liked, err := p.CheckIfPostLiked(ctx, postPostgres.Id.Bytes, uid)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to check if post %v is liked by user %v: %s", postPostgres.Id, uid, err.Error()))
			return nil, fmt.Errorf("unable to check if post is liked by user: %w", err)
		}
		postPostgres.IsLiked = pgtype.Bool{Bool: liked, Valid: true}

		result = append(result, postPostgres.ToPost())
	}

	return result, nil
}

func (p *PostgresPostRepository) UpdatePost(ctx context.Context, postUpdate models.PostUpdate) error {
	_, err := p.connPool.ExecContext(ctx, "update post set text = $1, updated_at = $2 where id = $3", postUpdate.Desc, time.Now(), postUpdate.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to update post %v in database: %s", postUpdate.Id, err.Error()))
		return fmt.Errorf("unable to update post in database: %w", err)
	}

	_, err = p.connPool.ExecContext(ctx, "delete from post_file where post_id = $1", postUpdate.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete post pictures %v from database: %s", postUpdate.Id, err.Error()))
		return fmt.Errorf("unable to delete post pictures from database: %w", err)
	}

	for _, file := range postUpdate.Files {
		_, err = p.connPool.ExecContext(ctx, insertPhotoQuery, postUpdate.Id, file.URL, file.DisplayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to insert post picture %v into database: %s", file.URL, err.Error()))
			return fmt.Errorf("unable to insert post picture into database: %w", err)
		}
	}

	return nil
}

func (p *PostgresPostRepository) GetPostFiles(ctx context.Context, postId uuid.UUID) ([]string, error) {
	rows, err := p.connPool.QueryContext(ctx, getPhotosQuery, postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get post pictures %v from database: %s", postId, err.Error()))
		return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var pic, tp, name pgtype.Text
		err = rows.Scan(&pic, &tp, &name)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan post picture %v from database: %s", pic, err.Error()))
			return nil, fmt.Errorf("unable to get post pictures from database: %w", err)
		}

		result = append(result, pic.String)
	}

	return result, nil
}

func (p *PostgresPostRepository) CheckIfPostLiked(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (bool, error) {
	var exists bool
	err := p.connPool.QueryRowContext(ctx, checkIfPostLikedRequest, postId, userId).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to check if post %v is liked by user %v: %s", postId, userId, err.Error()))
		return false, fmt.Errorf("unable to check if post is liked by user: %w", err)
	}
	return true, nil
}

func (p *PostgresPostRepository) UnlikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	res, err := p.connPool.ExecContext(ctx, unlikePostRequest, postId, userId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to unlike post %v by user %v: %s", postId, userId, err.Error()))
		return fmt.Errorf("unable to unlike post: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to check rows affected when unliking post %v by user %v: %s", postId, userId, err.Error()))
		return fmt.Errorf("unable to check if unlike was successful: %w", err)
	}

	if rowsAffected == 0 {
		return post_errors.ErrNotFound
	}

	return nil
}

func (p *PostgresPostRepository) LikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	// дополнительная проверка на лайк для безопасности
	liked, err := p.CheckIfPostLiked(ctx, postId, userId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to check if post %v is already liked by user %v: %s", postId, userId, err.Error()))
		return fmt.Errorf("failed to check if already liked: %w", err)
	}
	if liked {
		return post_errors.ErrAlreadyExists
	}

	_, err = p.connPool.ExecContext(ctx, likePostRequest, userId, postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to like post %v by user %v: %s", postId, userId, err.Error()))
		return fmt.Errorf("unable to like post: %w", err)
	}

	return nil
}
