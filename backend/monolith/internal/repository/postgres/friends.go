package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"quickflow/monolith/internal/models"
	postgresModels "quickflow/monolith/internal/repository/postgres/postgres-models"
	"quickflow/monolith/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
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
			where (user1_id = $1 or user2_id = $1) and status = $4
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
		where (user1_id = $1 and user2_id = $2) or (user1_id = $2 and user2_id = $1)
	`

	UpdateFriendRequestQuery = `
		update friendship
		set status = $3
		where user1_id = $1 and user2_id = $2 and status != $3
	`

	UpdateFriendStatusQuery = `
		update friendship
		set status = $3
		where user1_id = $1 and user2_id = $2 and status = $4
	`

	DeleteFollowerRelationQuery = `
		delete from friendship
		where ((user1_id = $1 and user2_id = $2) or (user1_id = $2 and user2_id = $1)) and status in ($3, $4)
	`

	GetFriendsCountQuery = `
	select count(*) 
	from friendship 
	where (user1_id = $1 or user2_id = $1) and status = $2
	`
)

type PostgresFriendsRepository struct {
	connPool *sql.DB
}

// NewPostgresFriendsRepository NewPostgresUserRepository creates new storage instance.
func NewPostgresFriendsRepository(db *sql.DB) *PostgresFriendsRepository {
	return &PostgresFriendsRepository{connPool: db}
}

// Close закрывает пул соединений
func (p *PostgresFriendsRepository) Close() {
	p.connPool.Close()
}

// GetFriendsPublicInfo Отдает структуру с информацией по друзьям + флаг hasMore, который говорит - остались ли еще друзья + ошибку
func (p *PostgresFriendsRepository) GetFriendsPublicInfo(ctx context.Context, userID string, limit int, offset int) ([]models.FriendInfo, bool, int, error) {
	logger.Info(ctx, fmt.Sprintf("Trying to get friends info for user %s", userID))

	rows, err := p.connPool.QueryContext(ctx, GetFriendsInfoQuery, userID, limit+1, offset, models.RelationFriend)
	friendsInfo := make([]models.FriendInfo, 0)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Error(ctx, newErr.Error())
		}

		return friendsInfo, false, 0, fmt.Errorf("unable to get friends info: %v", err)
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
			return []models.FriendInfo{}, false, 0, errors.New("unable to get friends info")
		}

		friendInfo := friendInfoPostgres.ConvertToFriendInfo()
		friendsInfo = append(friendsInfo, friendInfo)
	}

	var hasMore = false
	if len(friendsInfo) > limit {
		hasMore = true
		friendsInfo = friendsInfo[:limit]
	}

	logger.Info(ctx, fmt.Sprintf("Trying to get total amount of friends for user: %s", userID))

	var friendsCount int
	err = p.connPool.QueryRowContext(ctx, GetFriendsCountQuery, userID, models.RelationFriend).Scan(&friendsCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info(ctx, fmt.Sprintf("user: %s has no friends", userID))
			return []models.FriendInfo{}, false, 0, nil
		}
		logger.Error(ctx, fmt.Sprintf("unable to get friends count: %v", err))
		return []models.FriendInfo{}, false, 0, errors.New("unable to get friends info")
	}

	logger.Info(ctx, fmt.Sprintf("Amount of friends for user: %s is %d", userID, friendsCount))

	return friendsInfo, hasMore, friendsCount, nil
}

func (p *PostgresFriendsRepository) SendFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to insert friend request to DB for sender: %s and receiver %s", senderID, receiverID))
	var sender, receiver string
	var status models.UserRelation
	if senderID > receiverID {
		status = models.RelationFollowedBy
		receiver = senderID
		sender = receiverID
	} else {
		status = models.RelationFollowing
		receiver = receiverID
		sender = senderID
	}

	_, err := p.connPool.ExecContext(ctx, InsertFriendRequestQuery, sender, receiver, status)
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
	var status models.UserRelation

	err := p.connPool.QueryRowContext(ctx, CheckFriendRequestQuery, senderID, receiverID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	commandTag, err := p.connPool.ExecContext(ctx, UpdateFriendRequestQuery, sender, receiver, models.RelationFriend)
	if err != nil {
		return err
	}

	if rows, err := commandTag.RowsAffected(); rows == 0 || err != nil {
		logger.Error(ctx, fmt.Sprintf("friend relation between sender: %s and receiver: %s doesn't exist or incorrect ID's were given", senderID, receiverID))
		return errors.New("failed to accept friend request")
	}

	return nil
}

func (p *PostgresFriendsRepository) DeleteFriend(ctx context.Context, userID string, friendID string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to delete friend: %s for user: %s ", friendID, userID))
	var user1, user2 string
	var status models.UserRelation
	if userID < friendID {
		status = models.RelationFollowedBy
		user1 = userID
		user2 = friendID
	} else {
		status = models.RelationFollowing
		user1 = friendID
		user2 = userID
	}

	commandTag, err := p.connPool.ExecContext(ctx, UpdateFriendStatusQuery, user1, user2, status, models.RelationFriend)
	if err != nil {
		return err
	}

	if rows, err := commandTag.RowsAffected(); rows == 0 || err != nil {
		logger.Error(ctx, fmt.Sprintf("friend relation between sender: %s and receiver: %s doesn't exist or incorrect ID's were given", userID, friendID))
		return errors.New("failed to delete friend")
	}

	return nil
}

func (p *PostgresFriendsRepository) Unfollow(ctx context.Context, userID string, friendID string) error {
	logger.Info(ctx, fmt.Sprintf("Trying to unfollow user: %s for user: %s ", friendID, userID))

	commandTag, err := p.connPool.ExecContext(ctx, DeleteFollowerRelationQuery, userID, friendID, models.RelationFollowedBy, models.RelationFollowing)
	if err != nil {
		return err
	}

	if rows, err := commandTag.RowsAffected(); rows == 0 || err != nil {
		logger.Error(ctx, fmt.Sprintf("follower relation between user: %s and user: %s doesn't exist or incorrect ID's were given", userID, friendID))
		return errors.New("failed to delete friend")
	}

	return nil
}

func (p *PostgresFriendsRepository) GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error) {
	var status models.UserRelation
	err := p.connPool.QueryRowContext(ctx, CheckFriendRequestQuery, user1, user2).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RelationStranger, nil
		}
		logger.Error(ctx, fmt.Sprintf("unable to get friends info: %v", err))
		return models.RelationStranger, errors.New("unable to get friends info")
	}

	logger.Info(ctx, fmt.Sprintf("Relation between sender: %s and receiver: %s already exists", user1, user2))

	if user1.String() > user2.String() {
		if status == models.RelationFollowedBy {
			status = models.RelationFollowing
		} else if status == models.RelationFollowing {
			status = models.RelationFollowedBy
		}
	}
	return status, nil
}
