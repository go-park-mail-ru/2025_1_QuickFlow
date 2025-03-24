package redis

import (
	"context"
	"errors"
	"fmt"
	redis2 "quickflow/config/redis"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quickflow/internal/models"
)

type RedisSessionRepository struct {
	rdb *redis.Client
}

func NewRedisSessionRepository() *RedisSessionRepository {
	return &RedisSessionRepository{
		rdb: redis.NewClient(&redis.Options{
			Addr: redis2.NewRedisConfig().GetURL(),
		}),
	}
}

func (r *RedisSessionRepository) SaveSession(ctx context.Context, userId uuid.UUID, session models.Session) error {
	if err := r.rdb.Set(ctx, session.SessionId.String(), userId.String(), time.Until(session.ExpireDate)).Err(); err != nil {
		return fmt.Errorf("saving session error: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error) {
	userId, err := r.rdb.Get(ctx, session.SessionId.String()).Result()
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
	_, err := r.rdb.Get(ctx, session.String()).Result()

	switch {

	case errors.Is(err, redis.Nil):
		return false, nil

	case err != nil:
		return false, fmt.Errorf("redis connection failed: %w", err)

	default:
		return true, nil
	}
}

func (r *RedisSessionRepository) DeleteSession(ctx context.Context, session string) error {
	if err := r.rdb.Del(ctx, session).Err(); err != nil {
		return fmt.Errorf("unable to delete session: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) Close() {
	r.rdb.Close()
}
