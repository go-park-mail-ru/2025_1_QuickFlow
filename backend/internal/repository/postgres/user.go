package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config/postgres"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/internal/usecase"
	"quickflow/utils/validation"
)

const (
	insertUserQuery = `
	insert into "user" (id, username, psw_hash, salt)
	values ($1, $2, $3, $4)
`

	getUserByUsername = `
	select id, username, psw_hash, salt
	from "user" 
	where username = $1
`

	getUserByUIdQuery = `
	select id, username, psw_hash, salt 
	from "user"
	where id = $1
`

	searchSimilarUsersQuery = `
	select u.id, u.username, firstname, lastname, profile_avatar
	from profile p 
	join 
	(
	    select id, username, similarity(lower(username), lower($1)) as sim_factor
	    from "user"
	) u on p.id = u.id
	where u.sim_factor > 0.3
	order by sim_factor desc
	limit $2;
`
)

type PostgresUserRepository struct {
	connPool *pgxpool.Pool
}

// NewPostgresUserRepository creates new storage instance.
func NewPostgresUserRepository() *PostgresUserRepository {
	connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &PostgresUserRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (u *PostgresUserRepository) Close() {
	u.connPool.Close()
}

// IsExists checks if user with login exists.
func (u *PostgresUserRepository) IsExists(ctx context.Context, login string) (bool, error) {
	var id uuid.UUID

	err := u.connPool.QueryRow(ctx, "select id from \"user\" where username = $1", login).Scan(&id)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// SaveUser saves user to the repository.
func (u *PostgresUserRepository) SaveUser(ctx context.Context, user models.User) (uuid.UUID, error) {

	userPostgres := pgmodels.ConvertUserToPostgres(user)

	_, err := u.connPool.Exec(ctx, insertUserQuery,
		userPostgres.Id, userPostgres.Username,
		userPostgres.Password, userPostgres.Salt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to save user to database: %w", err)
	}

	return user.Id, nil
}

// GetUser returns user by login and password.
func (u *PostgresUserRepository) GetUser(ctx context.Context, loginData models.LoginData) (models.User, error) {
	var userPostgres pgmodels.UserPostgres

	err := u.connPool.QueryRow(ctx, getUserByUsername, loginData.Login).Scan(
		&userPostgres.Id, &userPostgres.Username,
		&userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	if !validation.CheckPassword(loginData.Password, userPostgres.Password.String, userPostgres.Salt.String) {
		return models.User{}, errors.New("incorrect login or password")
	}

	return userPostgres.ConvertToUser(), nil
}

// GetUserByUId returns user by id.
func (u *PostgresUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models.User, error) {
	var userPostgres pgmodels.UserPostgres

	err := u.connPool.QueryRow(ctx, getUserByUIdQuery,
		userId).Scan(&userPostgres.Id, &userPostgres.Username,
		&userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return userPostgres.ConvertToUser(), nil
}

func (u *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user pgmodels.UserPostgres
	err := u.connPool.QueryRow(ctx, getUserByUsername, username).Scan(&user.Id, &user.Username, &user.Password, &user.Salt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, usecase.ErrNotFound
	} else if err != nil {
		return models.User{}, err
	}

	return user.ConvertToUser(), nil
}

func (u *PostgresUserRepository) SearchSimilar(ctx context.Context, username string, postsCount uint) ([]models.PublicUserInfo, error) {
	rows, err := u.connPool.Query(ctx, searchSimilarUsersQuery, username, postsCount)
	if err != nil {
		return nil, fmt.Errorf("u.connPool.Query: %w", err)
	}
	defer rows.Close()

	var users []models.PublicUserInfo
	for rows.Next() {
		var user pgmodels.PublicUserInfoPostgres
		err = rows.Scan(&user.Id, &user.Username, &user.Firstname, &user.Lastname, &user.AvatarURL)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		users = append(users, user.ConvertToPublicUserInfo())
	}

	return users, nil
}
