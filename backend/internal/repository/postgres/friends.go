package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config/postgres"
	"quickflow/internal/models"
	postgresModels "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/pkg/logger"
)

const (
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
	where u.id = any (
    	select 
        	case 
            	when user1_id = $1 then user2_id
            	else user1_id 
        	end
    	from friendship
    	where user1_id = $1 or user2_id = $1
	)
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

func (p *PostgresFriendsRepository) GetFriendsPublicInfo(ctx context.Context, userID string) ([]models.FriendInfo, error) {
	logger.Info(ctx, fmt.Sprintf("Trying to get friends info for user %s", userID))

	rows, err := p.connPool.Query(ctx, GetFriendsInfoQuery, userID)
	defer rows.Close()
	friendsInfo := make([]models.FriendInfo, 0)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		}

		return friendsInfo, errors.New("unable to get friends info")
	}

	for rows.Next() {
		var friendInfoPostgres postgresModels.FriendInfoPostgres
		err = rows.Scan(
			&friendInfoPostgres.Id,
			&friendInfoPostgres.Username,
			&friendInfoPostgres.Firstname,
			&friendInfoPostgres.Lastname,
			&friendInfoPostgres.AvatarURL,
			&friendInfoPostgres.University,
		)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("rows scanning error: %s", err.Error()))
			return []models.FriendInfo{}, errors.New("unable to get friends info")
		}

		friendInfo := friendInfoPostgres.ConvertToFriendInfoPostgres()
		friendsInfo = append(friendsInfo, friendInfo)
	}

	logger.Info(ctx, friendsInfo)

	return friendsInfo, nil
}
