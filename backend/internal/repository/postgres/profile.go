package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"quickflow/config/postgres"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
)

const InsertProfileQuery = `
    insert into profile (id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date) 
        values ($1, $2, $3, $4, $5, $6, $7, $8);
`

const GetProfileQuery = `
    select id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date, school_id, university_id, contact_info_id
    from profile
    where id = $1;
`

const UpdateProfileQuery = `
    update profile 
    set bio = $2, profile_avatar = $3, profile_background = $4, firstname = $5, lastname = $6, 
        sex = $7, birth_date = $8
    where id = $1;
`

const UpdateContactInfoQuery = `
    update contact_info
    set city = $2, phone_number = $3, email = $4
    where id = (select contact_info_id from profile where id = $1);
`

const UpdateSchoolQuery = `
    update school
    set city = $2, name = $3
    where id = (select school_id from profile where id = $1);
`

const UpdateUniversityQuery = `
    update university
    set city = $2, name = $3, faculty = $4, graduation_year = $5
    where id = (select university_id from profile where id = $1);
`

const GetSchoolQuery = `
	select city, name
	from school
	where id = $1;
`

const GetUniversityQuery = `
	select name, city, faculty, graduation_year
	from university
	where id = $1;
`

const GetContactInfoQuery = `
	select city, email, phone_number
	from contact_info
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
	_, err := p.connPool.Exec(ctx, InsertProfileQuery, profile.UserId, profile.Bio, profile.AvatarUrl, profile.BackgroundUrl,
		profile.Name, profile.Surname, profile.Sex, profile.DateOfBirth)
	if err != nil {
		return fmt.Errorf("unable to save profile: %w", err)
	}

	return nil
}

func (p *PostgresProfileRepository) GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error) {
	var schoolId, universityId, contactInfoId pgtype.Int4

	var profile pgmodels.ProfilePostgres
	err := p.connPool.QueryRow(ctx, GetProfileQuery, userId).Scan(&profile.Id, &profile.Bio, &profile.AvatarUrl,
		&profile.BackgroundUrl, &profile.Name, &profile.Surname, &profile.Sex, &profile.DateOfBirth, &schoolId, &universityId,
		&contactInfoId)
	if err != nil {
		return models.Profile{}, fmt.Errorf("unable to get profile: %w", err)
	}

	if schoolId.Valid {
		profile.SchoolEducation = &pgmodels.SchoolEducation{}
		err := p.connPool.QueryRow(ctx, GetSchoolQuery, schoolId).Scan(&profile.SchoolEducation.City, &profile.SchoolEducation.School)
		if err != nil {
			return models.Profile{}, fmt.Errorf("unable to get school: %w", err)
		}
	}

	if universityId.Valid {
		profile.UniversityEducation = &pgmodels.UniversityEducation{}
		err := p.connPool.QueryRow(ctx, GetUniversityQuery, universityId).Scan(&profile.UniversityEducation.University,
			&profile.UniversityEducation.City, &profile.UniversityEducation.Faculty, &profile.UniversityEducation.GraduationYear)
		if err != nil {
			return models.Profile{}, fmt.Errorf("unable to get university: %w", err)
		}
	}

	if contactInfoId.Valid {
		profile.ContactInfo = &pgmodels.ContactInfoPostgres{}
		err := p.connPool.QueryRow(ctx, GetContactInfoQuery, contactInfoId).Scan(&profile.ContactInfo.City, &profile.ContactInfo.Email,
			&profile.ContactInfo.Phone)
		if err != nil {
			return models.Profile{}, fmt.Errorf("unable to get contact info: %w", err)
		}
	}

	return profile.ConvertToProfile(), nil
}

func (p *PostgresProfileRepository) UpdateProfile(ctx context.Context, newProfile models.Profile) error {
	tx, err := p.connPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// Обновляем основной профиль
	commandTag, err := tx.Exec(ctx, UpdateProfileQuery, newProfile.UserId, newProfile.Bio, newProfile.AvatarUrl,
		newProfile.BackgroundUrl, newProfile.Name, newProfile.Surname, newProfile.Sex, newProfile.DateOfBirth)
	if err != nil {
		return fmt.Errorf("unable to update profile: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("profile not found")
	}

	var contactInfoID, schoolID, universityID pgtype.Int4

	// 🔹 Проверяем и обновляем ContactInfo
	if newProfile.ContactInfo != nil {
		err = tx.QueryRow(ctx, `SELECT id FROM contact_info WHERE phone_number = $1 AND email = $2`,
			newProfile.ContactInfo.Phone, newProfile.ContactInfo.Email).Scan(&contactInfoID)

		if err == pgx.ErrNoRows {
			// Вставляем новую запись
			err = tx.QueryRow(ctx, `
				INSERT INTO contact_info (city, phone_number, email)
				VALUES ($1, $2, $3)
				RETURNING id;
			`, newProfile.ContactInfo.City, newProfile.ContactInfo.Phone, newProfile.ContactInfo.Email).Scan(&contactInfoID)
		} else if err == nil {
			// Обновляем существующую запись
			_, err = tx.Exec(ctx, `
				UPDATE contact_info 
				SET city = $1 
				WHERE id = $2;
			`, newProfile.ContactInfo.City, contactInfoID)
		}
		if err != nil {
			return fmt.Errorf("unable to update contact info: %w", err)
		}
	}

	// 🔹 Проверяем и обновляем School
	if newProfile.SchoolEducation != nil {
		err = tx.QueryRow(ctx, `SELECT id FROM school WHERE city = $1 AND name = $2`,
			newProfile.SchoolEducation.City, newProfile.SchoolEducation.School).Scan(&schoolID)

		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `
				INSERT INTO school (city, name)
				VALUES ($1, $2)
				RETURNING id;
			`, newProfile.SchoolEducation.City, newProfile.SchoolEducation.School).Scan(&schoolID)
		}
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("unable to update school education: %w", err)
		}
	}

	// 🔹 Проверяем и обновляем University
	if newProfile.UniversityEducation != nil {
		err = tx.QueryRow(ctx, `SELECT id FROM university WHERE name = $1 AND faculty = $2 AND graduation_year = $3`,
			newProfile.UniversityEducation.University, newProfile.UniversityEducation.Faculty, newProfile.UniversityEducation.GraduationYear).Scan(&universityID)

		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `
				INSERT INTO university (city, name, faculty, graduation_year)
				VALUES ($1, $2, $3, $4)
				RETURNING id;
			`, newProfile.UniversityEducation.City, newProfile.UniversityEducation.University,
				newProfile.UniversityEducation.Faculty, newProfile.UniversityEducation.GraduationYear).Scan(&universityID)
		}
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("unable to update university education: %w", err)
		}
	}

	// 🔹 Обновляем ссылки на внешние ключи в профиле
	if newProfile.SchoolEducation != nil || newProfile.UniversityEducation != nil || newProfile.ContactInfo != nil {
		_, err = tx.Exec(ctx, `
			UPDATE profile 
			SET contact_info_id = $2, school_id = $3, university_id = $4
			WHERE id = $1;
		`, newProfile.UserId, contactInfoID, schoolID, universityID)
		if err != nil {
			return fmt.Errorf("unable to update profile relations: %w", err)
		}
	}

	return nil
}
