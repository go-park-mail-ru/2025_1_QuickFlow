package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"quickflow/internal/models"
	postgres_models "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/internal/usecase"
	"quickflow/pkg/logger"
	"time"
)

const (
	saveFeedbackQuery = `
	insert into feedback (rating, respondent_id, text, type, created_at) 
	values ($1, $2, $3, $4, $5)
`

	getFeedbackOlderQuery = `
	select rating, respondent_id, text, type, created_at
	from feedback
	where created_at < $1 and type = $2
	order by created_at desc
	limit $3
`

	getAverageRatingTypeQuery = `
		select avg(rating)
		from feedback
		where type = $1;
`

	getAverateRatingQuery = `
	select type, avg(rating)
	from feedback
	group by type
	`

	getNumMessagesSent = `
	select count
	from count_messages
	where user_id = $1;
	`

	getNumPostsCreated = `
	select count
	from count_post
	where user_id = $1;
	`

	getNumProfileChanges = `
	select count
	from count_profile
	where user_id = $1;
	`
)

type FeedbackRepository struct {
	ConnPool *sql.DB
}

func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{ConnPool: db}
}

// Close закрывает пул соединений
func (f *FeedbackRepository) Close() {
	f.ConnPool.Close()
}

func (f *FeedbackRepository) SaveFeedback(ctx context.Context, feedback *models.Feedback) error {
	pgFeedback := postgres_models.FromModel(feedback)
	_, err := f.ConnPool.
		ExecContext(ctx, saveFeedbackQuery, pgFeedback.Rating, pgFeedback.RespondentId, pgFeedback.Text, pgFeedback.Type, pgFeedback.CreatedAt)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to save feedback: %v", err))
		return fmt.Errorf("save feedback: %w", err)
	}

	return nil
}

func (f *FeedbackRepository) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	rows, err := f.ConnPool.QueryContext(ctx, getFeedbackOlderQuery,
		pgtype.Timestamptz{Time: ts, Valid: true}, pgtype.Text{String: string(feedbackType), Valid: true}, count)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, usecase.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
		return nil, fmt.Errorf("get feedback: %w", err)
	}
	defer rows.Close()
	var feedback []models.Feedback
	for rows.Next() {
		var r postgres_models.PgFeedback
		err = rows.Scan(&r.Rating, &r.RespondentId, &r.Text, &r.Type, &r.CreatedAt)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
			return nil, fmt.Errorf("get feedback: %w", err)
		}

		feedback = append(feedback, r.ToModel())
	}

	return feedback, nil
}

func (f *FeedbackRepository) GetAverageRatingType(ctx context.Context, feedbackType models.FeedbackType) (float64, error) {
	var avg float64
	err := f.ConnPool.QueryRowContext(ctx, getAverageRatingTypeQuery, feedbackType).Scan(&avg)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, usecase.ErrNotFound
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
		return 0, fmt.Errorf("get feedback: %w", err)
	}

	return avg, nil
}

func (f *FeedbackRepository) GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error) {
	var num int64
	err := f.ConnPool.QueryRowContext(ctx, getNumMessagesSent, userId).Scan(&num)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
		return 0, fmt.Errorf("get feedback: %w", err)
	}
	return num, nil
}

func (f *FeedbackRepository) GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error) {
	var num int64
	err := f.ConnPool.QueryRowContext(ctx, getNumPostsCreated, userId).Scan(&num)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
		return 0, fmt.Errorf("get feedback: %w", err)
	}
	return num, nil
}

func (f *FeedbackRepository) GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error) {
	var num int64
	err := f.ConnPool.QueryRowContext(ctx, getNumProfileChanges, userId).Scan(&num)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get feedback: %v", err))
		return 0, fmt.Errorf("get feedback: %w", err)
	}
	return num, nil
}
