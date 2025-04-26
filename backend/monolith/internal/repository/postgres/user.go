package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	models2 "quickflow/monolith/internal/models"
	"quickflow/monolith/internal/repository/postgres/postgres-models"
	"quickflow/monolith/internal/usecase"
	"quickflow/monolith/utils/validation"

	"github.com/google/uuid"
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
	SELECT id, username, firstname, lastname, profile_avatar
	FROM (
		SELECT u.id,
			   username,
			   firstname,
			   lastname,
			   profile_avatar,
			   similarity(lower(username), lower($1)) AS sim_factor_username,
			   similarity(lower(firstname || ' ' || lastname), lower($1)) AS sim_factor_full_name
		FROM "user" u JOIN profile p ON u.id = p.id
	) t
	WHERE GREATEST(t.sim_factor_username, t.sim_factor_full_name) > 0.3
	ORDER BY GREATEST(t.sim_factor_username, t.sim_factor_full_name) DESC
	LIMIT $2;
`
)

type PostgresUserRepository struct {
	connPool *sql.DB
}

// NewPostgresUserRepository creates new storage instance.
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{connPool: db}
}

// Close закрывает пул соединений
func (u *PostgresUserRepository) Close() {
	u.connPool.Close()
}

// IsExists checks if user with login exists.
func (u *PostgresUserRepository) IsExists(ctx context.Context, login string) (bool, error) {
	var id uuid.UUID

	err := u.connPool.QueryRowContext(ctx, "select id from \"user\" where username = $1", login).Scan(&id)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// SaveUser saves user to the repository.
func (u *PostgresUserRepository) SaveUser(ctx context.Context, user models2.User) (uuid.UUID, error) {

	userPostgres := postgres_models.ConvertUserToPostgres(user)

	_, err := u.connPool.ExecContext(ctx, insertUserQuery,
		userPostgres.Id, userPostgres.Username,
		userPostgres.Password, userPostgres.Salt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to save user to database: %w", err)
	}

	return user.Id, nil
}

// GetUser returns user by login and password.
func (u *PostgresUserRepository) GetUser(ctx context.Context, loginData models2.LoginData) (models2.User, error) {
	var userPostgres postgres_models.UserPostgres

	err := u.connPool.QueryRowContext(ctx, getUserByUsername, loginData.Login).Scan(
		&userPostgres.Id, &userPostgres.Username,
		&userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models2.User{}, errors.New("user not found")
	}

	if !validation.CheckPassword(loginData.Password, userPostgres.Password.String, userPostgres.Salt.String) {
		return models2.User{}, errors.New("incorrect login or password")
	}

	return userPostgres.ConvertToUser(), nil
}

// GetUserByUId returns user by id.
func (u *PostgresUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models2.User, error) {
	var userPostgres postgres_models.UserPostgres

	err := u.connPool.QueryRowContext(ctx, getUserByUIdQuery,
		userId).Scan(&userPostgres.Id, &userPostgres.Username,
		&userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models2.User{}, errors.New("user not found")
	}

	return userPostgres.ConvertToUser(), nil
}

func (u *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (models2.User, error) {
	var user postgres_models.UserPostgres
	err := u.connPool.QueryRowContext(ctx, getUserByUsername, username).Scan(&user.Id, &user.Username, &user.Password, &user.Salt)
	if errors.Is(err, sql.ErrNoRows) {
		return models2.User{}, usecase.ErrNotFound
	} else if err != nil {
		return models2.User{}, err
	}

	return user.ConvertToUser(), nil
}

func (u *PostgresUserRepository) SearchSimilar(ctx context.Context, toSearch string, postsCount uint) ([]models2.PublicUserInfo, error) {
	rows, err := u.connPool.QueryContext(ctx, searchSimilarUsersQuery, toSearch, postsCount)
	if err != nil {
		return nil, fmt.Errorf("u.connPool.Query: %w", err)
	}
	defer rows.Close()

	var users []models2.PublicUserInfo
	for rows.Next() {
		var user postgres_models.PublicUserInfoPostgres
		err = rows.Scan(&user.Id, &user.Username, &user.Firstname, &user.Lastname, &user.AvatarURL)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		users = append(users, user.ConvertToPublicUserInfo())
	}

	return users, nil
}
