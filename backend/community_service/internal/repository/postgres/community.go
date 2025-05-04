package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	community_errors "quickflow/community_service/internal/errors"
	postgres_models "quickflow/community_service/internal/repository/postgres-models"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

const (
	createCommunityQuery = `
	insert into community (id, owner_id, name, description, created_at, avatar_url)
	values ($1, $2, $3, $4, $5, $6);
`

	getCommunityByIdQuery = `
	select id, owner_id, name, description, created_at, avatar_url
	from community
	where id = $1;
`

	getCommunityByNameQuery = `
	select id, owner_id, name, description, created_at, avatar_url
	from community
	where name = $1;
`

	getCommunityMembersQuery = `
	select user_id, community_id, role, joined_at
	from community_user
	where community_id = $1 and joined_at < $3
	order by joined_at desc
	limit $2;
`

	isCommunityMemberQuery = `
	select id, role
	from community_user
	where user_id = $1 and community_id = $2;
`

	deleteCommunityQuery = `
	delete from community
	where id = $1;
`

	updateCommunityQuery = `
	update community
	set name = $1, description = $2, avatar_url = $3
	where id = $4;
`

	joinCommunityRequest = `
	insert into community_user (user_id, community_id, role, joined_at)
	values ($1, $2, $3, $4);
`

	leaveCommunity = `
	delete from community_user where user_id = $1 and community_id = $2;
`

	getUserCommunities = `
	select c.id, c.owner_id, c.name, c.description, c.created_at, c.avatar_url
	from community c
	join community_user cu on c.id = cu.community_id
	where cu.user_id = $1 and cu.joined_at < $3
	order by cu.joined_at desc
	limit $2;
`

	searchSimilarCommunities = `
	SELECT c.id, c.owner_id, c.name, c.description, c.created_at, c.avatar_url
	FROM (
		SELECT id, owner_id, name, description, created_at, avatar_url,
			   similarity(lower(name), lower($1)) AS sim_factor_name
		FROM community
	) c
	WHERE c.sim_factor_name > 0.3
	ORDER BY sim_factor_name DESC
	LIMIT $2;
`
)

type SqlCommunityRepository struct {
	connPool *sql.DB
}

func NewSqlCommunityRepository(connPool *sql.DB) *SqlCommunityRepository {
	return &SqlCommunityRepository{
		connPool: connPool,
	}
}

// Close закрывает пул соединений
func (c *SqlCommunityRepository) Close() {
	c.connPool.Close()
}

func (c *SqlCommunityRepository) CreateCommunity(ctx context.Context, community models.Community) error {
	communityDTO := postgres_models.CommunityModelToPostgres(&community)
	_, err := c.connPool.ExecContext(ctx, createCommunityQuery,
		communityDTO.Id, communityDTO.OwnerId, communityDTO.Name,
		communityDTO.Description, communityDTO.CreatedAt, communityDTO.AvatarUrl)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to create community: %v", err))
		return err
	}

	// add owner as admin
	_, err = c.connPool.ExecContext(ctx, joinCommunityRequest,
		communityDTO.OwnerId, communityDTO.Id,
		string(models.CommunityRoleOwner), communityDTO.CreatedAt)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to add owner as admin: %v", err))
		return err
	}

	return nil
}

func (c *SqlCommunityRepository) GetCommunityById(ctx context.Context, id uuid.UUID) (models.Community, error) {
	communityDTO := postgres_models.CommunityPostgres{}
	err := c.connPool.QueryRowContext(ctx, getCommunityByIdQuery, id).Scan(
		&communityDTO.Id, &communityDTO.OwnerId, &communityDTO.Name,
		&communityDTO.Description, &communityDTO.CreatedAt, &communityDTO.AvatarUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Community{}, community_errors.ErrNotFound
		}
		logger.Error(ctx, fmt.Sprintf("unable to get community by id: %v", err))
		return models.Community{}, err
	}

	return communityDTO.PostgresToCommunityModel(), nil
}

func (c *SqlCommunityRepository) GetCommunityByName(ctx context.Context, name string) (models.Community, error) {
	communityDTO := postgres_models.CommunityPostgres{}
	err := c.connPool.QueryRowContext(ctx, getCommunityByNameQuery, name).Scan(
		&communityDTO.Id, &communityDTO.OwnerId, &communityDTO.Name,
		&communityDTO.Description, &communityDTO.CreatedAt, &communityDTO.AvatarUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Community{}, community_errors.ErrNotFound
		}
		logger.Error(ctx, fmt.Sprintf("unable to get community by id: %v", err))
		return models.Community{}, err
	}

	return communityDTO.PostgresToCommunityModel(), nil
}

func (c *SqlCommunityRepository) GetCommunityMembers(ctx context.Context, id uuid.UUID, numMembers int, ts time.Time) ([]models.CommunityMember, error) {
	rows, err := c.connPool.QueryContext(ctx, getCommunityMembersQuery, id, numMembers, ts)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Info(ctx, "no community members found")
		return nil, community_errors.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to get community members: %v", err))
		return nil, err
	}
	defer rows.Close()

	var members []models.CommunityMember
	for rows.Next() {
		var member postgres_models.CommunityMemberPostgres
		if err := rows.Scan(&member.UserId, &member.CommunityId, &member.Role, &member.JoinedAt); err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to scan community member: %v", err))
			return nil, err
		}
		members = append(members, member.PostgresToCommunityMemberModel())
	}

	if err := rows.Err(); err != nil {
		logger.Error(ctx, fmt.Sprintf("error iterating over community members: %v", err))
		return nil, err
	}

	return members, nil
}

func (c *SqlCommunityRepository) IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error) {
	if userId == uuid.Nil || communityId == uuid.Nil {
		logger.Error(ctx, "user ID and community ID cannot be empty")
		return false, nil, fmt.Errorf("user ID and community ID cannot be empty")
	}

	var id pgtype.Int4
	var role string
	err := c.connPool.QueryRowContext(ctx, isCommunityMemberQuery, userId, communityId).Scan(&id, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		logger.Error(ctx, fmt.Sprintf("unable to check if user is a community member: %v", err))
		return false, nil, err
	}

	rl := models.CommunityRole(role)
	return true, &rl, nil
}
func (c *SqlCommunityRepository) DeleteCommunity(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		logger.Error(ctx, "community ID cannot be empty")
		return fmt.Errorf("community ID cannot be empty")
	}

	_, err := c.connPool.ExecContext(ctx, deleteCommunityQuery, id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to delete community: %v", err))
		return err
	}

	return nil
}

func (c *SqlCommunityRepository) UpdateCommunity(ctx context.Context, community models.Community) error {
	res, err := c.connPool.ExecContext(ctx, updateCommunityQuery,
		community.Name, community.Description, community.AvatarUrl, community.ID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to update community: %v", err))
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to get affected rows: %v", err))
		return err
	}
	if affected == 0 {
		logger.Error(ctx, "no rows affected")
		return community_errors.ErrNotFound
	}

	return nil
}

func (c *SqlCommunityRepository) JoinCommunity(ctx context.Context, member models.CommunityMember) error {
	memberPostgres := postgres_models.CommunityMemberModelToPostgres(&member)
	_, err := c.connPool.ExecContext(ctx, joinCommunityRequest,
		memberPostgres.UserId, memberPostgres.CommunityId,
		memberPostgres.Role, memberPostgres.JoinedAt)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to join community: %v", err))
		return err
	}

	return nil
}

func (c *SqlCommunityRepository) LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error {
	_, err := c.connPool.ExecContext(ctx, leaveCommunity,
		pgtype.UUID{Bytes: userId, Valid: true}, pgtype.UUID{Bytes: communityId, Valid: true})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to leave community: %v", err))
		return err
	}

	return nil
}

func (c *SqlCommunityRepository) GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error) {
	rows, err := c.connPool.QueryContext(ctx, getUserCommunities, userId, count, pgtype.Timestamptz{Time: ts, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		logger.Info(ctx, "no user communities found")
		return nil, community_errors.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to get user communities by id: %v", err))
		return nil, err
	}
	defer rows.Close()

	var communities []models.Community
	for rows.Next() {
		var communityDTO postgres_models.CommunityPostgres
		if err := rows.Scan(&communityDTO.Id, &communityDTO.OwnerId, &communityDTO.Name,
			&communityDTO.Description, &communityDTO.CreatedAt, &communityDTO.AvatarUrl); err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to scan user community: %v", err))
			return nil, err
		}
		communities = append(communities, communityDTO.PostgresToCommunityModel())
	}

	return communities, nil
}

func (c *SqlCommunityRepository) SearchSimilarCommunities(ctx context.Context, name string, count int) ([]models.Community, error) {
	rows, err := c.connPool.QueryContext(ctx, searchSimilarCommunities, name, count)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("unable to search similar communities: %v", err))
		return nil, err
	}
	defer rows.Close()

	var communities []models.Community
	for rows.Next() {
		var communityDTO postgres_models.CommunityPostgres
		if err := rows.Scan(&communityDTO.Id, &communityDTO.OwnerId, &communityDTO.Name,
			&communityDTO.Description, &communityDTO.CreatedAt, &communityDTO.AvatarUrl); err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to scan similar community: %v", err))
			return nil, err
		}
		communities = append(communities, communityDTO.PostgresToCommunityModel())
	}
	if err := rows.Err(); err != nil {
		logger.Error(ctx, fmt.Sprintf("error iterating over similar communities: %v", err))
		return nil, err
	}

	return communities, nil
}
