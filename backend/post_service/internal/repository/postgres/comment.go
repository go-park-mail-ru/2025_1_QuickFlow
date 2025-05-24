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
	postgres_models "quickflow/post_service/internal/repository/postgres-models"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

const getCommentQuery = `
	select id, post_id, user_id, text, created_at, updated_at, like_count
	from comment
	where id = $1
`

const getCommentsForPostQuery = `
    select id, post_id, user_id, text, created_at, updated_at, like_count
    from comment
    where post_id = $1 and created_at < $2
    order by created_at desc
    limit $3;
`

const getCommentFilesQuery = `
	select file_url, file_type
	from comment_file
	where comment_id = $1
	order by added_at;
`

const insertCommentQuery = `
	insert into comment (id, post_id, user_id, created_at, text, like_count)
	values ($1, $2, $3, $4, $5, $6)
`

const insertCommentFileQuery = `
	insert into comment_file (comment_id, file_url, file_type)
	values ($1, $2, $3)
`

const checkIfCommentLikedRequest = `
	select 1
	from like_comment
	where comment_id = $1 and user_id = $2;
`

const likeCommentRequest = `
	insert into like_comment (user_id, comment_id)
	values ($1, $2);
`

const unlikeCommentRequest = `
	delete from like_comment
	where comment_id = $1 and user_id = $2;
`

const getLastCommentQuery = `
	select id, post_id, user_id, text, created_at, updated_at, like_count
	from comment
	where post_id = $1 and created_at = (select max(created_at) from comment where post_id = $1) limit 1;
`

type PostgresCommentRepository struct {
	connPool *sql.DB
}

func NewPostgresCommentRepository(connPool *sql.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{
		connPool: connPool,
	}
}

// AddComment adds a comment to the repository.
func (c *PostgresCommentRepository) AddComment(ctx context.Context, comment models.Comment) error {
	// Преобразуем в Postgres-совместимую модель
	commentPostgres := postgres_models.ConvertCommentToPostgres(comment)
	_, err := c.connPool.ExecContext(ctx, insertCommentQuery,
		commentPostgres.Id, commentPostgres.PostId, commentPostgres.UserId,
		commentPostgres.CreatedAt, commentPostgres.Text, commentPostgres.LikeCount)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to save comment %v to database: %s", comment, err.Error()))
		return fmt.Errorf("unable to save comment to database: %w", err)
	}

	// Сохраняем файлы (включая display_type, который хранится в file_type)
	for _, file := range commentPostgres.Files {
		_, err = c.connPool.ExecContext(ctx, insertCommentFileQuery, commentPostgres.Id, file.URL, file.DisplayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to save comment files %v for comment %v: %s", commentPostgres.Files, comment, err.Error()))
			return fmt.Errorf("unable to save comment files to database: %w", err)
		}
	}
	return nil
}

// GetCommentFiles получает файлы, связанные с комментарием.
func (c *PostgresCommentRepository) GetCommentFiles(ctx context.Context, commentId uuid.UUID) ([]string, error) {
	rows, err := c.connPool.QueryContext(ctx, getCommentFilesQuery, commentId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get comment files %v from database: %s", commentId, err.Error()))
		return nil, fmt.Errorf("unable to get comment files from database: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var file pgtype.Text
		var displayType pgtype.Text // Для получения типа отображаемого файла
		err = rows.Scan(&file, &displayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan comment file %v from database: %s", file, err.Error()))
			return nil, fmt.Errorf("unable to scan comment file from database: %w", err)
		}
		// Здесь мы можем использовать displayType, если необходимо
		result = append(result, file.String)
	}

	return result, nil
}

// DeleteComment удаляет комментарий из репозитория.
func (c *PostgresCommentRepository) DeleteComment(ctx context.Context, commentId uuid.UUID) error {
	_, err := c.connPool.ExecContext(ctx, "delete from comment cascade where id = $1", pgtype.UUID{Bytes: commentId, Valid: true})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete comment %v from database: %s", commentId, err.Error()))
		return fmt.Errorf("unable to delete comment from database: %w", err)
	}
	return nil
}

// GetCommentsForPost получает комментарии для поста из репозитория.
func (c *PostgresCommentRepository) GetCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error) {
	rows, err := c.connPool.QueryContext(ctx, getCommentsForPostQuery, postId, timestamp, numComments)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, post_errors.ErrNotFound
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get comments from database for post %v, numComments %v, timestamp %v: %s", postId, numComments, timestamp, err.Error()))
		return nil, fmt.Errorf("unable to get comments from database: %w", err)
	}
	defer rows.Close()

	var result []models.Comment
	for rows.Next() {
		var commentPostgres postgres_models.CommentPostgres
		err = rows.Scan(&commentPostgres.Id, &commentPostgres.PostId, &commentPostgres.UserId, &commentPostgres.Text,
			&commentPostgres.CreatedAt, &commentPostgres.UpdatedAt, &commentPostgres.LikeCount)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan comment %v from database: %s", commentPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get comments from database: %w", err)
		}

		// Получаем файлы для комментария, включая display_type
		files, err := c.connPool.QueryContext(ctx, getCommentFilesQuery, commentPostgres.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to get comment files %v from database: %s", commentPostgres.Id, err.Error()))
			return nil, fmt.Errorf("unable to get comment files from database: %w", err)
		}

		isLiked, err := c.CheckIfCommentLiked(ctx, commentPostgres.Id.Bytes, commentPostgres.UserId.Bytes)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to check if comment %v is liked by user %v: %s", commentPostgres.Id, commentPostgres.UserId, err.Error()))
			return nil, fmt.Errorf("unable to check if comment is liked by user: %w", err)
		}
		commentPostgres.IsLiked = pgtype.Bool{Bool: isLiked, Valid: true}

		for files.Next() {
			var pic pgtype.Text
			var displayType pgtype.Text
			err = files.Scan(&pic, &displayType)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan comment file %v from database: %s", pic, err.Error()))
				return nil, fmt.Errorf("unable to get comment files from database: %w", err)
			}

			commentPostgres.Files = append(commentPostgres.Files, &postgres_models.PostgresFile{URL: pic, DisplayType: displayType})
		}
		files.Close()

		result = append(result, commentPostgres.ToComment())
	}

	return result, nil
}

// GetComment получает комментарий по ID.
func (c *PostgresCommentRepository) GetComment(ctx context.Context, commentId uuid.UUID) (models.Comment, error) {
	row := c.connPool.QueryRowContext(ctx, getCommentQuery, commentId)
	var commentPostgres postgres_models.CommentPostgres
	err := row.Scan(&commentPostgres.Id, &commentPostgres.PostId, &commentPostgres.UserId, &commentPostgres.Text,
		&commentPostgres.CreatedAt, &commentPostgres.UpdatedAt, &commentPostgres.LikeCount)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get comment %v from database: %s", commentId, err.Error()))
		return models.Comment{}, fmt.Errorf("unable to get comment from database: %w", err)
	}

	files, err := c.connPool.QueryContext(ctx, getCommentFilesQuery, commentPostgres.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get comment files %v from database: %s", commentPostgres.Id, err.Error()))
		return models.Comment{}, fmt.Errorf("unable to get comment files from database: %w", err)
	}

	for files.Next() {
		var pic pgtype.Text
		var displayType pgtype.Text
		err = files.Scan(&pic, &displayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan comment file %v from database: %s", pic, err.Error()))
			return models.Comment{}, fmt.Errorf("unable to get comment files from database: %w", err)
		}

		commentPostgres.Files = append(commentPostgres.Files, &postgres_models.PostgresFile{URL: pic, DisplayType: displayType})
	}
	files.Close()

	isLiked, err := c.CheckIfCommentLiked(ctx, commentPostgres.Id.Bytes, commentPostgres.UserId.Bytes)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to check if comment %v is liked by user %v: %s", commentPostgres.Id, commentPostgres.UserId, err.Error()))
		return models.Comment{}, fmt.Errorf("unable to check if comment is liked by user: %w", err)
	}
	commentPostgres.IsLiked = pgtype.Bool{Bool: isLiked, Valid: true}

	return commentPostgres.ToComment(), nil
}

// UpdateComment обновляет комментарий и добавляет новые файлы.
func (c *PostgresCommentRepository) UpdateComment(ctx context.Context, commentUpdate models.CommentUpdate) error {
	// Обновление текста комментария
	_, err := c.connPool.ExecContext(ctx, "update comment set text = $1, updated_at = $2 where id = $3", commentUpdate.Text, pgtype.Timestamptz{Time: time.Now(), Valid: true}, pgtype.UUID{Bytes: commentUpdate.Id, Valid: true})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to update comment %v in database: %s", commentUpdate.Id, err.Error()))
		return fmt.Errorf("unable to update comment in database: %w", err)
	}

	// Добавление новых файлов
	for _, file := range commentUpdate.Files {
		if file == nil {
			continue
		}
		filePostgres := postgres_models.FileToPostgres(*file)
		_, err = c.connPool.ExecContext(ctx, insertCommentFileQuery, commentUpdate.Id, filePostgres.URL, filePostgres.DisplayType) // Используем DisplayType
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to add file %v for comment %v: %s", file.URL, commentUpdate.Id, err.Error()))
			return fmt.Errorf("unable to add file to comment: %w", err)
		}
	}

	return nil
}

// CheckIfCommentLiked проверяет, лайкнул ли пользователь комментарий.
func (c *PostgresCommentRepository) CheckIfCommentLiked(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (bool, error) {
	var exists bool
	err := c.connPool.QueryRowContext(ctx, checkIfCommentLikedRequest, commentId, userId).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		// Если в таблице нет записи, то возвращаем false
		return false, nil
	} else if err != nil {
		// Ошибка при выполнении запроса
		logger.Error(ctx, fmt.Sprintf("Unable to check if comment %v is liked by user %v: %s", commentId, userId, err.Error()))
		return false, fmt.Errorf("unable to check if comment is liked by user: %w", err)
	}

	return exists, nil
}

// LikeComment ставит лайк на комментарий.
func (c *PostgresCommentRepository) LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
	// Проверяем, не поставил ли уже пользователь лайк
	liked, err := c.CheckIfCommentLiked(ctx, commentId, userId)
	if err != nil {
		return fmt.Errorf("failed to check if comment is liked: %w", err)
	}

	if liked {
		// Если лайк уже поставлен, ничего не делаем (идемпотентность)
		return nil
	}

	// Вставляем лайк в таблицу like_comment
	_, err = c.connPool.ExecContext(ctx, likeCommentRequest, userId, commentId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to like comment %v by user %v: %s", commentId, userId, err.Error()))
		return fmt.Errorf("unable to like comment: %w", err)
	}
	return nil
}

// UnlikeComment убирает лайк с комментария.
func (c *PostgresCommentRepository) UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
	// Проверяем, поставил ли пользователь лайк
	liked, err := c.CheckIfCommentLiked(ctx, commentId, userId)
	if err != nil {
		return fmt.Errorf("failed to check if comment is liked: %w", err)
	}

	if !liked {
		// Если лайк не поставлен, ничего не делаем (идемпотентность)
		return nil
	}

	// Удаляем лайк из таблицы like_comment
	_, err = c.connPool.ExecContext(ctx, unlikeCommentRequest, commentId, userId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to unlike comment %v by user %v: %s", commentId, userId, err.Error()))
		return fmt.Errorf("unable to unlike comment: %w", err)
	}
	return nil
}

func (c *PostgresCommentRepository) GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error) {
	// Получаем комментарий
	row := c.connPool.QueryRowContext(ctx, getLastCommentQuery, postId)
	var commentPostgres postgres_models.CommentPostgres
	err := row.Scan(&commentPostgres.Id, &commentPostgres.PostId, &commentPostgres.UserId, &commentPostgres.Text,
		&commentPostgres.CreatedAt, &commentPostgres.UpdatedAt, &commentPostgres.LikeCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, post_errors.ErrNotFound
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get comment %v from database: %s", postId, err.Error()))
		return nil, fmt.Errorf("unable to get comment from database: %w", err)
	}

	isLiked, err := c.CheckIfCommentLiked(ctx, commentPostgres.Id.Bytes, commentPostgres.UserId.Bytes)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to check if comment %v is liked by user %v: %s", commentPostgres.Id, commentPostgres.UserId, err.Error()))
		return nil, fmt.Errorf("unable to check if comment is liked by user: %w", err)
	}
	commentPostgres.IsLiked = pgtype.Bool{Bool: isLiked, Valid: true}

	res := commentPostgres.ToComment()
	return &res, nil
}
