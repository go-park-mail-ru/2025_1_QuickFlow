package factory

import (
	"database/sql"
	"quickflow/monolith/config"

	"quickflow/monolith/internal/repository/minio"
	postgres2 "quickflow/monolith/internal/repository/postgres"
	"quickflow/monolith/internal/repository/redis"
	usecase2 "quickflow/monolith/internal/usecase"
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

func (f *PGMFactory) UserRepository() usecase2.UserRepository {
	return postgres2.NewPostgresUserRepository(f.db)
}

func (f *PGMFactory) PostRepository() usecase2.PostRepository {
	return postgres2.NewPostgresPostRepository(f.db)
}

func (f *PGMFactory) SessionRepository() usecase2.SessionRepository {
	return f.redisRepo
}

func (f *PGMFactory) ProfileRepository() usecase2.ProfileRepository {
	return postgres2.NewPostgresProfileRepository(f.db)
}

func (f *PGMFactory) ChatRepository() usecase2.ChatRepository {
	return postgres2.NewPostgresChatRepository(f.db)
}

func (f *PGMFactory) MessageRepository() usecase2.MessageRepository {
	return postgres2.NewPostgresMessageRepository(f.db)
}

func (f *PGMFactory) FileRepository() usecase2.FileRepository {
	return f.minioRepo
}

func (f *PGMFactory) FriendRepository() usecase2.FriendsRepository {
	return postgres2.NewPostgresFriendsRepository(f.db)
}

func (f *PGMFactory) Close() error {
	if err := f.db.Close(); err != nil {
		return err
	}
	return nil
}
