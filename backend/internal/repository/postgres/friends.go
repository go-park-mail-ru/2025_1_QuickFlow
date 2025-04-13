package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
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
		with friends as (
			select 
				case 
					when user1_id = $1 then user2_id
					else user1_id 
				end as friend_id
			from friendship
			where (user1_id = $1 or user2_id = $1) and status = 'friend'
		)
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
		where u.id in (select friend_id from friends)
		order by p.lastname, p.firstname
		limit $2
		offset $3
	`

	InsertFriendRequestQuery = `
		insert into friendship (user1_id, user2_id, status)
		values ($1, $2, $3)
	`

	CheckFriendRequestQuery = `
		select status
		from friendship
		where user1_id = $1 and user2_id = $2
	`

	UpdateFriendRequestQuery = `
		update friendship
		set status = 'friend'
		where user1_id = $1 and user2_id = $2 and status != 'friend'
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

// GetFriendsPublicInfo Отдает структуру с информацией по друзьям + флаг hasMore, который говорит - остались ли еще друзья + ошибку
func (p *PostgresFriendsRepository) GetFriendsPublicInfo(ctx context.Context, userID string, limit int, offset int) ([]models.FriendInfo, bool, error) {
	logger.Info(ctx, fmt.Sprintf("Trying to get friends info for user %s", userID))

	rows, err := p.connPool.Query(ctx, GetFriendsInfoQuery, userID, limit+1, offset)
	defer rows.Close()
	friendsInfo := make([]models.FriendInfo, 0)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		}

		return friendsInfo, false, fmt.Errorf("unable to get friends info: %v", err)
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
			return []models.FriendInfo{}, false, errors.New("unable to get friends info")
		}

		friendInfo := friendInfoPostgres.ConvertToFriendInfo()
		friendsInfo = append(friendsInfo, friendInfo)
	}

	var hasMore = false
	if len(friendsInfo) > limit {
		hasMore = true
		friendsInfo = friendsInfo[:limit]
	}

	return friendsInfo, hasMore, nil
}

func (p *PostgresFriendsRepository) SendFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to insert friend request to DB for sender: %s and receiver %s", senderID, receiverID))
	var status, sender, receiver string
	if senderID > receiverID {
		status = "followed_by"
		receiver = senderID
		sender = receiverID
	} else {
		status = "following"
		receiver = receiverID
		sender = senderID
	}

	_, err := p.connPool.Exec(ctx, InsertFriendRequestQuery, sender, receiver, status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		}

		return fmt.Errorf("unable to get friends info: %v", err)
	}
	return nil
}

func (p *PostgresFriendsRepository) IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error) {
	var status, sender, receiver string
	if senderID > receiverID {
		receiver = senderID
		sender = receiverID
	} else {
		receiver = receiverID
		sender = senderID
	}

	err := p.connPool.QueryRow(ctx, CheckFriendRequestQuery, sender, receiver).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info(ctx, fmt.Sprintf("relation between sender: %s and receiver: %s doesn't exist or Incorrect IDs were given", senderID, receiverID))
			return false, nil
		}
		logger.Error(ctx, fmt.Sprintf("unable to get friends info: %v", err))
		return false, errors.New("unable to get friends info")
	}

	logger.Error(ctx, fmt.Sprintf("Relation between sender: %s and receiver: %s already exists", senderID, receiverID))
	return true, nil
}

func (p *PostgresFriendsRepository) AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to update friend request for sender: %s and receiver: %s", senderID, receiverID))
	var sender, receiver string
	if senderID > receiverID {
		receiver = senderID
		sender = receiverID
	} else {
		receiver = receiverID
		sender = senderID
	}

	commandTag, err := p.connPool.Exec(ctx, UpdateFriendRequestQuery, sender, receiver)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		logger.Error(ctx, fmt.Sprintf("friend relation between sender: %s and receiver: %s doesn't exist or incorrect ID's were given", senderID, receiverID))
		return errors.New("failed to accept friend request")
	}

	return nil
}
