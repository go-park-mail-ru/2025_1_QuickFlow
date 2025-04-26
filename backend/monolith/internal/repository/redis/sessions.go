package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	redis2 "quickflow/monolith/config/redis"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/pkg/logger"
)

type RedisSessionRepository struct {
	rdb *redis.Client
}

func NewRedisSessionRepository() *RedisSessionRepository {
	redisCfg := redis2.NewRedisConfig()

	return &RedisSessionRepository{
		rdb: redis.NewClient(&redis.Options{
			Addr:     redisCfg.GetURL(),
			Password: redisCfg.GetPass(),
		}),
	}
}

func (r *RedisSessionRepository) SaveSession(ctx context.Context, userId uuid.UUID, session models.Session) error {
	logger.Info(ctx, fmt.Sprintf("Trying to save session in Redis for userId: %s", userId.String()))

	if err := r.rdb.Set(ctx, session.SessionId.String(), userId.String(), time.Until(session.ExpireDate)).Err(); err != nil {
		logger.Error(ctx, "Failed to save session to redis")
		return fmt.Errorf("saving session error: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("Successfully saved session in Redis for userId: %s", userId.String()))

	return nil
}

func (r *RedisSessionRepository) LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error) {
	logger.Info(ctx, fmt.Sprintf("Trying to get user in Redis for sessionId: %s", session.SessionId.String()))

	userId, err := r.rdb.Get(ctx, session.SessionId.String()).Result()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to find user in Redis for sessionId: %s", session.SessionId.String()))
		return uuid.Nil, fmt.Errorf("unable to get userId: %w", err)
	}

	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		logger.Error(ctx, "Failed to parse userId from Redis")
		return uuid.Nil, fmt.Errorf("unable to parse userId: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("Successfully found user in Redis for sessionId: %s", session.SessionId.String()))

	return userIdUUID, nil
}

func (r *RedisSessionRepository) IsExists(ctx context.Context, session uuid.UUID) (bool, error) {
	logger.Info(ctx, fmt.Sprintf("Trying to find user in Redis for sessionId: %s", session.String()))

	_, err := r.rdb.Get(ctx, session.String()).Result()

	switch {

	case errors.Is(err, redis.Nil):
		logger.Info(ctx, fmt.Sprintf("User does not exist in Redis for sessionId: %s", session.String()))
		return false, nil

	case err != nil:
		logger.Error(ctx, fmt.Sprintf("Redis connection failed: %s", err.Error()))
		return false, fmt.Errorf("failed to find user in Redis for sessionId: %w", err)

	default:
		logger.Info(ctx, fmt.Sprintf("Successfully found user in Redis for sessionId: %s", session.String()))
		return true, nil
	}
}

func (r *RedisSessionRepository) DeleteSession(ctx context.Context, session string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to delete session in Redis for sessionId: %s", session))

	if err := r.rdb.Del(ctx, session).Err(); err != nil {
		logger.Error(ctx, fmt.Sprintf("Redis connection failed: %s", err.Error()))
		return fmt.Errorf("unable to delete session: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("Successfully deleted session in Redis for sessionId: %s", session))

	return nil
}

func (r *RedisSessionRepository) Close() {
	err := r.rdb.Close()
	if err != nil {
		log.Fatal("unable to close Redis connection:", err.Error())
	}
}
