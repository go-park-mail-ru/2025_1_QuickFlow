package factory

import (
	"database/sql"

	"quickflow/config"
	"quickflow/internal/repository/minio"
	"quickflow/internal/repository/postgres"
	"quickflow/internal/repository/redis"
	"quickflow/internal/usecase"
)

type PGMFactory struct {
	db        *sql.DB
	minioRepo *minio.MinioRepository
	redisRepo *redis.RedisSessionRepository
}

func NewPGMFactory(cfg *config.Config) (*PGMFactory, error) {
	db, err := sql.Open("pgx", cfg.PostgresConfig.GetURL())
	if err != nil {
		return nil, err
	}
	fileRepo, err := minio.NewMinioRepository(cfg.MinioConfig)
	if err != nil {
		return nil, err
	}
	redisRepo := redis.NewRedisSessionRepository()

	return &PGMFactory{
		db:        db,
		minioRepo: fileRepo,
		redisRepo: redisRepo,
	}, nil
}

func (f *PGMFactory) UserRepository() usecase.UserRepository {
	return postgres.NewPostgresUserRepository(f.db)
}

func (f *PGMFactory) PostRepository() usecase.PostRepository {
	return postgres.NewPostgresPostRepository(f.db)
}

func (f *PGMFactory) SessionRepository() usecase.SessionRepository {
	return f.redisRepo
}

func (f *PGMFactory) ProfileRepository() usecase.ProfileRepository {
	return postgres.NewPostgresProfileRepository(f.db)
}

func (f *PGMFactory) ChatRepository() usecase.ChatRepository {
	return postgres.NewPostgresChatRepository(f.db)
}

func (f *PGMFactory) MessageRepository() usecase.MessageRepository {
	return postgres.NewPostgresMessageRepository(f.db)
}

func (f *PGMFactory) FileRepository() usecase.FileRepository {
	return f.minioRepo
}

func (f *PGMFactory) FriendRepository() usecase.FriendsRepository {
	return postgres.NewPostgresFriendsRepository(f.db)
}

func (f *PGMFactory) Close() error {
	if err := f.db.Close(); err != nil {
		return err
	}
	return nil
}
