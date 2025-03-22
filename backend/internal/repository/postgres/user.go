package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"quickflow/config/postgres"
	"quickflow/utils/validation"

	"github.com/google/uuid"

	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
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
func (p *PostgresUserRepository) Close() {
	p.connPool.Close()
}

// IsExists checks if user with login exists.
func (i *PostgresUserRepository) IsExists(ctx context.Context, login string) (bool, error) {
	var id int64

	err := i.connPool.QueryRow(ctx, "select id from users where login = $1", login).Scan(&id)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// SaveUser saves user to the repository.
func (i *PostgresUserRepository) SaveUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	userPostgres := pgmodels.ConvertUserToPostgres(user)

	_, err := i.connPool.Exec(ctx, insertUserQuery,
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
	var userPostgres pgmodels.UserPostgres

	err := i.connPool.QueryRow(ctx, getUserByLogin, loginData.Login).Scan(
		&userPostgres.Id, &userPostgres.Name,
		&userPostgres.Surname, &userPostgres.Sex,
		&userPostgres.DateOfBirth, &userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	if !validation.CheckPassword(loginData.Password, userPostgres.Password.String, userPostgres.Salt.String) {
		return models.User{}, errors.New("incorrect login or password")
	}

	return userPostgres.ConvertToUser(), nil
}

// GetUserByUId returns user by id.
func (i *PostgresUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models.User, error) {
	var userPostgres pgmodels.UserPostgres

	err := i.connPool.QueryRow(ctx, getUserByUIdQuery,
		userId).Scan(&userPostgres.Id, &userPostgres.Name,
		&userPostgres.Surname, &userPostgres.Sex,
		&userPostgres.DateOfBirth, &userPostgres.Password, &userPostgres.Salt)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return userPostgres.ConvertToUser(), nil
}
