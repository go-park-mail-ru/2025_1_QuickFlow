package postgres_redis

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"

    "quickflow/config"
    "quickflow/internal/models"
    "quickflow/utils"
)

type PostgresUserRepository struct {
}

func (i *PostgresUserRepository) IsExists(ctx context.Context, login string) bool {
    conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
    if err != nil {
        // TODO: return uuid.Nil, fmt.Errorf("unable to connect to database: %w", err)
        return false
    }
    defer conn.Close(ctx)

    var id int64

    err = conn.QueryRow(ctx, "select id from users where login = $1",
        login).Scan(&id)
    if err != nil {
        return false
    }

    return true
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

    _, err = conn.Exec(ctx, "insert into users (uuid, login, name, surname, sex, birth_date, psw_hash, salt) values ($1, $2, $3, $4, $5, $6, $7, $8)",
        user.Id, user.Login, user.Name, user.Surname, user.Sex, user.DateOfBirth, user.Password, user.Salt)
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

    var (
        userID       uuid.UUID
        name         string
        surname      string
        sex          models.Sex
        dateOfBirth  pgtype.Date
        passwordHash string
        salt         string
    )

    err = conn.QueryRow(ctx, "select uuid, name, surname, sex, birth_date, psw_hash, salt from users where login = $1",
        loginData.Login).Scan(&userID, &name, &surname, &sex, &dateOfBirth, &passwordHash, &salt)
    if err != nil {
        return models.User{}, errors.New("user not found")
    }

    if !utils.CheckPassword(loginData.Password, passwordHash, salt) {
        return models.User{}, errors.New("incorrect login or password")
    }

    return models.User{
        Id:          userID,
        Login:       loginData.Login,
        Name:        name,
        Surname:     surname,
        Password:    passwordHash,
        DateOfBirth: dateOfBirth.Time.String(),
        Sex:         sex,
        Salt:        salt,
    }, nil
}

// GetUserByUId returns user by id.
func (i *PostgresUserRepository) GetUserByUId(ctx context.Context, userId uuid.UUID) (models.User, error) {
    conn, err := pgx.Connect(ctx, config.NewPostgresConfig().GetURL())
    if err != nil {
        return models.User{}, fmt.Errorf("unable to connect to database: %w", err)
    }
    defer conn.Close(ctx)

    var (
        name         string
        surname      string
        sex          models.Sex
        dateOfBirth  pgtype.Date
        passwordHash string
        salt         string
        login        string
    )

    err = conn.QueryRow(ctx, "select login, name, surname, sex, birth_date, psw_hash, salt from users where uuid = $1",
        userId).Scan(&login, &name, &surname, &sex, &dateOfBirth, &passwordHash, &salt)
    if err != nil {
        return models.User{}, errors.New("user not found")
    }

    return models.User{
        Id:          userId,
        Login:       login,
        Name:        name,
        Surname:     surname,
        Password:    passwordHash,
        DateOfBirth: dateOfBirth.Time.String(),
        Sex:         sex,
        Salt:        salt,
    }, nil
}
