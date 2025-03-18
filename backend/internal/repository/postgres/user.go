package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"quickflow/config"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/utils"
)

const insertUserQuery = `
	insert into users (uuid, login, name, surname, sex, birth_date, psw_hash, salt)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
`

const getUserByLogin = `
	select uuid, name, surname, sex, birth_date, psw_hash, salt
	from users 
	where login = $1
`

const getUserByUIdQuery = `
	select uuid, name, surname, sex, birth_date, psw_hash, salt 
	from users
	where uuid = $1
`

type PostgresUserRepository struct {
}

// IsExists checks if user with login exists.
func (i *PostgresUserRepository) IsExists(ctx context.Context, login string) (bool, error) {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return false, fmt.Errorf("unable to connect to database: %w", err)

	}
	defer conn.Close(ctx)

	var id int64

	err = conn.QueryRow(ctx, "select id from users where login = $1", login).Scan(&id)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// NewPostgresUserRepository creates new storage instance.
func NewPostgresUserRepository() *PostgresUserRepository {
	return &PostgresUserRepository{}
}

// SaveUser saves user to the repository.
func (i *PostgresUserRepository) SaveUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	userPostgres := pgmodels.ConvertUserToPostgres(user)

	_, err = conn.Exec(ctx, insertUserQuery,
		userPostgres.Id, userPostgres.Login, userPostgres.Name,
		userPostgres.Surname, userPostgres.Sex, userPostgres.DateOfBirth,
		userPostgres.Password, userPostgres.Salt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to save user to database: %w", err)
	}

	return user.Id, nil
}

// GetUser returns user by login and password.
func (i *PostgresUserRepository) GetUser(ctx context.Context, loginData models.LoginData) (models.User, error) {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return models.User{}, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	var userPostgres pgmodels.UserPostgres

	err = conn.QueryRow(ctx, getUserByLogin, loginData.Login).Scan(
		&userPostgres.Id, &userPostgres.Name,
		&userPostgres.Surname, &userPostgres.Sex,
		&userPostgres.DateOfBirth, &userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	if !utils.CheckPassword(loginData.Password, userPostgres.Password.String, userPostgres.Salt.String) {
		return models.User{}, errors.New("incorrect login or password")
	}

	return userPostgres.ConvertToUser(), nil
}

// GetUserByUId returns user by id.
func (i *PostgresUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models.User, error) {
	conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
	if err != nil {
		return models.User{}, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	var userPostgres pgmodels.UserPostgres

	err = conn.QueryRow(ctx, getUserByUIdQuery,
		userId).Scan(&userPostgres.Id, &userPostgres.Name,
		&userPostgres.Surname, &userPostgres.Sex,
		&userPostgres.DateOfBirth, &userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return userPostgres.ConvertToUser(), nil
}
