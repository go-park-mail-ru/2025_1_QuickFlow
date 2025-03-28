package postgres

import (
    "context"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "log"
    "quickflow/config/postgres"
    "quickflow/internal/models"
    pgmodels "quickflow/internal/repository/postgres/postgres-models"
)

const InsertProfileQuery = `
    insert into profile (id, bio, profile_avatar, firstname, lastname, sex, birth_date) 
        values ($1, $2, $3, $4, $5, $6, $7);
`

const GetProfileQuery = `
    select id, bio, profile_avatar, firstname, lastname, sex, birth_date
    from profile
    where id = $1;
`

const UpdateProfileQuery = `
    update profile 
    set bio = $2, profile_avatar = $3, firstname = $4, lastname = $5, 
        sex = $6, birth_date = $7
    where id = $1;
`

type PostgresProfileRepository struct {
    connPool *pgxpool.Pool
}

func NewPostgresProfileRepository() *PostgresProfileRepository {
    connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
    if err != nil {
        log.Fatalf("Unable to create connection pool: %v", err)
    }

    return &PostgresProfileRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (p *PostgresProfileRepository) Close() {
    p.connPool.Close()
}

func (p *PostgresProfileRepository) SaveProfile(ctx context.Context, profile models.Profile) error {
    _, err := p.connPool.Exec(ctx, InsertProfileQuery, profile.UserId, profile.Bio, profile.AvatarUrl,
        profile.Name, profile.Surname, profile.Sex, profile.DateOfBirth)
    if err != nil {
        return fmt.Errorf("unable to save profile: %w", err)
    }

    return nil
}

func (p *PostgresProfileRepository) GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error) {
    var profile pgmodels.ProfilePostgres
    err := p.connPool.QueryRow(ctx, GetProfileQuery, userId).Scan(&profile.Id, &profile.Bio, &profile.AvatarUrl,
        &profile.Name, &profile.Surname, &profile.Sex, &profile.DateOfBirth)
    if err != nil {
        return models.Profile{}, fmt.Errorf("unable to get profile: %w", err)
    }

    return profile.ConvertToProfile(), nil
}

func (p *PostgresProfileRepository) UpdateProfile(ctx context.Context, newProfile models.Profile) error {
    commandTag, err := p.connPool.Exec(ctx, UpdateProfileQuery, newProfile.UserId, newProfile.Bio, newProfile.AvatarUrl,
        newProfile.Name, newProfile.Surname, newProfile.Sex, newProfile.DateOfBirth)
    if err != nil {
        return fmt.Errorf("unable to update profile: %w", err)
    }

    if commandTag.RowsAffected() == 0 {
        return errors.New("profile not found")
    }

    return nil
}
