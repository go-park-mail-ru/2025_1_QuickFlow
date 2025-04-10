package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"quickflow/internal/models"
	"quickflow/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config/postgres"
)

const (
	GetFriendIDsQuery = `
		select user2_id
		from friendship
		where user1_id = $1
	`

	GetFriendsInfoQuery = `
	select 
		u.id, 
		u.username, 
		p.firstname, 
		p.lastname, 
		p.profile_avatar, 
		univ.name
	from "user" u
	join profile p on u.id = p.id
	left join education e on e.profile_id = p.id
	left join faculty f on f.id = e.faculty_id
	left join university univ on f.university_id = univ.id
	where u.id = any($1)
`
)

type PostgresFriendsRepository struct {
	connPool *pgxpool.Pool
}

// NewPostgresFriendsRepository NewPostgresUserRepository creates new storage instance.
func NewPostgresFriendsRepository() *PostgresFriendsRepository {
	connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &PostgresFriendsRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (p *PostgresFriendsRepository) Close() {
	p.connPool.Close()
}

func (p *PostgresFriendsRepository) GetFriends(ctx context.Context, userId uuid.UUID) ([]string, error) {
	logger.Info(ctx, fmt.Sprintf("Getting friends for user %s", userId))

	rows, err := p.connPool.Query(ctx, GetFriendIDsQuery, userId)
	defer rows.Close()
	friendIds := make([]string, 0)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		} else {
			logger.Error(ctx, fmt.Sprintf("Some other error: %s", err.Error()))
		}

		return friendIds, errors.New("unable to get friends ids")
	}

	for rows.Next() {
		var friendID string
		err = rows.Scan(&friendID)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("rows scanning error: %s", err.Error()))
			return []string{}, errors.New("unable to get friends ids")
		}
		friendIds = append(friendIds, friendID)
	}

	return friendIds, nil
}

func (p *PostgresFriendsRepository) GetFriendsInfo(ctx context.Context, friendIDs []string) ([]models.FriendInfo, error) {
	logger.Info(ctx, "Trying to get friends info")

	rows, err := p.connPool.Query(ctx, GetFriendIDsQuery, friendIDs)
	defer rows.Close()
	friendsInfo := make([]models.FriendInfo, 0)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		} else {
			logger.Error(ctx, fmt.Sprintf("Some other error: %s", err.Error()))
		}

		return friendsInfo, errors.New("unable to get friends info")
	}

	for rows.Next() {
		var friendInfo models.FriendInfo
		err = rows.Scan(&friendInfo)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("rows scanning error: %s", err.Error()))
			return []models.FriendInfo{}, errors.New("unable to get friends info")
		}
		friendsInfo = append(friendsInfo, friendInfo)
	}

	return []models.FriendInfo{}, nil
}
