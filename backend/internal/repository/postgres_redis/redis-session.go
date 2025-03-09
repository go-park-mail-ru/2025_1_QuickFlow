package postgres_redis

import (
	"context"
	"errors"
	"fmt"
	"quickflow/config"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quickflow/internal/models"
)

type RedisSessionRepository struct {
}

func NewRedisSessionRepository() *RedisSessionRepository {
	return &RedisSessionRepository{}
}

func (r *RedisSessionRepository) SaveSession(ctx context.Context, userId uuid.UUID, session models.Session) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.NewRedisConfig().GetURL(),
	})

	defer rdb.Close()

	if err := rdb.Set(ctx, session.SessionId.String(), userId, time.Until(session.ExpireDate)).Err(); err != nil {
		return fmt.Errorf("saving session error: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.NewRedisConfig().GetURL(),
	})

	defer rdb.Close()

	userId, err := rdb.Get(ctx, session.SessionId.String()).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to get userId: %w", err)
	}

	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse userId: %w", err)
	}

	return userIdUUID, nil
}

func (r *RedisSessionRepository) IsExists(ctx context.Context, session uuid.UUID) (bool, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.NewRedisConfig().GetURL(),
	})

	defer rdb.Close()

	_, err := rdb.Get(ctx, session.String()).Result()

	switch {

	case errors.Is(err, redis.Nil):
		return false, nil

	case err != nil:
		return false, fmt.Errorf("redis conetction failed: %w", err)

	default:
		return true, nil
	}

}
