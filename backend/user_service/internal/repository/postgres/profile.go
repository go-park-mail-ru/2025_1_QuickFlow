package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/pkg/logger"
	"quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	pgmodels "quickflow/user_service/internal/repository/postgres-models"
)

const InsertProfileQuery = `
    insert into profile (id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date) 
        values ($1, $2, $3, $4, $5, $6, $7, $8);
`

const GetProfileQuery = `
    select id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date, school_id, contact_info_id, last_seen
    from profile
    where id = $1;
`

const UpdateProfileQuery = `
    update profile 
    set bio = $2, firstname = $3, lastname = $4, 
        sex = $5, birth_date = $6
    where id = $1;
`

const GetSchoolQuery = `
	select city, name
	from school
	where id = $1;
`

const GetContactInfoQuery = `
	select city, email, phone_number
	from contact_info
	where id = $1;
`

const InsertOrGetUniversityQuery = `
    insert into university (name, city)
    values ($1, $2)
    on conflict (name, city) do update set name = excluded.name
    returning id;
`

const InsertOrGetFacultyQuery = `
    insert into faculty (university_id, name)
    values ($1, $2)
    on conflict (university_id, name) do update set name = excluded.name
    returning id;
`

const InsertOrUpdateEducationQuery = `
    insert into education (profile_id, faculty_id, graduation_year)
    values ($1, $2, $3)
    on conflict (profile_id) do update 
    set faculty_id = excluded.faculty_id, graduation_year = excluded.graduation_year;
`

const GetEducationQuery = `
    select u.name, u.city, f.name, e.graduation_year
    from education e
        join faculty f on e.faculty_id = f.id
        join university u on f.university_id = u.id
    where e.profile_id = $1;
`

const GetPublicUserInfoQuery = `
	select u.id, firstname, lastname, profile_avatar, username, last_seen
	from profile p join "user" u on p.id = u.id
	where u.id = $1
`

const GetPublicUsersInfoQuery = `
	select u.id, firstname, lastname, profile_avatar, username, last_seen
	from profile p join "user" u on p.id = u.id
	where u.id = any($1)
`
const updateLastSeenQuery = `
	update profile
	set last_seen = $2
	where id = $1
`

type PostgresProfileRepository struct {
	connPool *sql.DB
}

// NewPostgresProfileRepository создает новый экземпляр репозитория.
func NewPostgresProfileRepository(db *sql.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{
		connPool: db,
	}
}

// Close закрывает пул соединений
func (p *PostgresProfileRepository) Close() {
	p.connPool.Close()
}

// SaveProfile сохраняет профиль пользователя в базе данных.
func (p *PostgresProfileRepository) SaveProfile(ctx context.Context, profile models.Profile) error {
	_, err := p.connPool.ExecContext(ctx, InsertProfileQuery, profile.UserId, profile.BasicInfo.Bio,
		profile.BasicInfo.AvatarUrl, profile.BasicInfo.BackgroundUrl,
		profile.BasicInfo.Name, profile.BasicInfo.Surname, profile.BasicInfo.Sex,
		profile.BasicInfo.DateOfBirth)
	if err != nil {
		return fmt.Errorf("unable to save profile: %w", err)
	}
	return nil
}

// GetProfile получает профиль пользователя из базы данных по его идентификатору.
func (p *PostgresProfileRepository) GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error) {
	var schoolId, contactInfoId pgtype.Int4

	var profile pgmodels.ProfilePostgres
	err := p.connPool.QueryRowContext(ctx, GetProfileQuery, userId).Scan(&profile.Id, &profile.Bio, &profile.AvatarUrl,
		&profile.BackgroundUrl, &profile.Name, &profile.Surname, &profile.Sex, &profile.DateOfBirth, &schoolId,
		&contactInfoId, &profile.LastSeen)
	if err != nil {
		return models.Profile{}, fmt.Errorf("unable to get profile: %w", err)
	}

	if schoolId.Valid {
		profile.SchoolEducation = &pgmodels.SchoolEducation{}
		err := p.connPool.QueryRowContext(ctx, GetSchoolQuery, schoolId).Scan(&profile.SchoolEducation.City, &profile.SchoolEducation.School)
		if err != nil {
			return models.Profile{}, fmt.Errorf("unable to get school: %w", err)
		}
	}

	profile.UniversityEducation = &pgmodels.UniversityEducation{}
	err = p.connPool.QueryRowContext(ctx, GetEducationQuery, userId).Scan(
		&profile.UniversityEducation.University,
		&profile.UniversityEducation.City,
		&profile.UniversityEducation.Faculty,
		&profile.UniversityEducation.GraduationYear,
	)
	if errors.Is(err, sql.ErrNoRows) {
		profile.UniversityEducation = nil
	} else if err != nil {
		return models.Profile{}, fmt.Errorf("unable to get university education: %w", err)
	}

	if contactInfoId.Valid {
		profile.ContactInfo = &pgmodels.ContactInfoPostgres{}
		err := p.connPool.QueryRowContext(ctx, GetContactInfoQuery, contactInfoId).Scan(&profile.ContactInfo.City, &profile.ContactInfo.Email,
			&profile.ContactInfo.Phone)
		if err != nil {
			return models.Profile{}, fmt.Errorf("unable to get contact info: %w", err)
		}
	}

	return profile.ConvertToProfile(), nil
}

// UpdateProfileTextInfo обновляет текстовую информацию профиля в базе данных.
func (p *PostgresProfileRepository) UpdateProfileTextInfo(ctx context.Context, newProfile models.Profile) error {
	tx, err := p.connPool.Begin()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to begin transaction: %v", err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Обновляем основной профиль
	if newProfile.BasicInfo != nil {
		commandTag, err := tx.ExecContext(ctx, UpdateProfileQuery, newProfile.UserId, newProfile.BasicInfo.Bio,
			newProfile.BasicInfo.Name, newProfile.BasicInfo.Surname, newProfile.BasicInfo.Sex,
			newProfile.BasicInfo.DateOfBirth)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update profile: %v", err))
			return fmt.Errorf("unable to update profile: %w", err)
		}
		if rows, err := commandTag.RowsAffected(); rows == 0 || err != nil {
			return user_errors.ErrNotFound
		}
	}

	// Обновляем username, если он был изменен
	if newProfile.Username != "" {
		_, err = tx.ExecContext(ctx, `update "user" set username = $1 where id = $2`, newProfile.Username, newProfile.UserId)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update username: %v", err))
			return fmt.Errorf("unable to update username: %w", err)
		}
	}

	var contactInfoID, schoolID pgtype.Int4
	// обновляем ContactInfo
	if newProfile.ContactInfo != nil {
		contactInfoID, err = updateContactInfo(ctx, tx, *newProfile.ContactInfo)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update contact info: %v", err))
			return fmt.Errorf("unable to update contact info: %w", err)
		}
	}

	// обновляем School
	if newProfile.SchoolEducation != nil {
		schoolID, err = updateSchoolInfo(ctx, tx, *newProfile.SchoolEducation)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update education: %v", err))
			return fmt.Errorf("unable to update school education: %w", err)
		}
	}

	// обновляем University
	if newProfile.UniversityEducation != nil {
		err = updateUniversityInfo(ctx, tx, newProfile.UserId, *newProfile.UniversityEducation)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update university education: %v", err))
			return fmt.Errorf("unable to update university education: %w", err)
		}
	}

	// Обновляем ссылки на внешние ключи в профиле
	if newProfile.SchoolEducation != nil || newProfile.UniversityEducation != nil || newProfile.ContactInfo != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE profile 
			SET contact_info_id = $2, school_id = $3
			WHERE id = $1;
		`, newProfile.UserId, contactInfoID, schoolID)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("unable to update profile relations to education tables: %v", err))
			return fmt.Errorf("unable to update profile relations: %w", err)
		}
	}

	return nil
}

// UpdateProfileAvatar обновляет аватар профиля в базе данных.
func (p *PostgresProfileRepository) UpdateProfileAvatar(ctx context.Context, id uuid.UUID, url string) error {
	_, err := p.connPool.QueryContext(ctx, `update profile set profile_avatar = $1 where id = $2`, url, id)
	if err != nil {
		return fmt.Errorf("unable to update profile avatar: %w", err)
	}
	return nil
}

// UpdateProfileCover обновляет обложку профиля в базе данных.
func (p *PostgresProfileRepository) UpdateProfileCover(ctx context.Context, id uuid.UUID, url string) error {
	_, err := p.connPool.QueryContext(ctx, `update profile set profile_background = $1 where id = $2`, url, id)
	if err != nil {
		return fmt.Errorf("unable to update profile background: %w", err)
	}
	return nil
}

func (p *PostgresProfileRepository) GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models.PublicUserInfo, error) {
	var publicInfo pgmodels.PublicUserInfoPostgres
	err := p.connPool.QueryRowContext(ctx, GetPublicUserInfoQuery, userId).Scan(
		&publicInfo.Id, &publicInfo.Firstname, &publicInfo.Lastname,
		&publicInfo.AvatarURL, &publicInfo.Username, &publicInfo.LastSeen)
	if err != nil {
		return models.PublicUserInfo{}, fmt.Errorf("unable to get public user info: %w", err)
	}
	return publicInfo.ConvertToPublicUserInfo(), nil
}

func (p *PostgresProfileRepository) GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]models.PublicUserInfo, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	rows, err := p.connPool.QueryContext(ctx, GetPublicUsersInfoQuery, userIds)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user_errors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("unable to get public user info: %w", err)
	}
	defer rows.Close()

	var publicInfos []models.PublicUserInfo
	for rows.Next() {
		var publicInfo pgmodels.PublicUserInfoPostgres
		err := rows.Scan(
			&publicInfo.Id, &publicInfo.Firstname, &publicInfo.Lastname,
			&publicInfo.AvatarURL, &publicInfo.Username, &publicInfo.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("unable to scan public user info: %w", err)
		}
		publicInfos = append(publicInfos, publicInfo.ConvertToPublicUserInfo())
	}

	return publicInfos, nil
}

func (p *PostgresProfileRepository) UpdateLastSeen(ctx context.Context, userId uuid.UUID) error {
	_, err := p.connPool.ExecContext(ctx, updateLastSeenQuery, userId, time.Now())
	if err != nil {
		return fmt.Errorf("u.connPool.Exec: %w", err)
	}
	return nil
}

func updateContactInfo(ctx context.Context, tx *sql.Tx, contactInfo models.ContactInfo) (pgtype.Int4, error) {
	var contactInfoID pgtype.Int4

	err := tx.QueryRowContext(ctx, `SELECT id FROM contact_info WHERE phone_number = $1 AND email = $2`,
		contactInfo.Phone, contactInfo.Email).Scan(&contactInfoID)

	if errors.Is(err, sql.ErrNoRows) {
		// Вставляем новую запись
		err = tx.QueryRowContext(ctx, `
			INSERT INTO contact_info (city, phone_number, email)
			VALUES ($1, $2, $3)
			RETURNING id;
		`, contactInfo.City, contactInfo.Phone, contactInfo.Email).Scan(&contactInfoID)
	} else if err == nil {
		// Обновляем существующую запись
		_, err = tx.ExecContext(ctx, `
			UPDATE contact_info 
			SET city = $1 
			WHERE id = $2;
		`, contactInfo.City, contactInfoID)
	}
	if err != nil {
		return pgtype.Int4{}, fmt.Errorf("unable to update contact info: %w", err)
	}

	return contactInfoID, nil
}

func updateSchoolInfo(ctx context.Context, tx *sql.Tx, education models.SchoolEducation) (pgtype.Int4, error) {
	var schoolID pgtype.Int4
	err := tx.QueryRowContext(ctx, `SELECT id FROM school WHERE city = $1 AND name = $2`,
		education.City, education.School).Scan(&schoolID)

	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
				INSERT INTO school (city, name)
				VALUES ($1, $2)
				RETURNING id;
			`, education.City, education.School).Scan(&schoolID)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return pgtype.Int4{}, fmt.Errorf("unable to update school education: %w", err)
	}
	return schoolID, nil
}

func updateUniversityInfo(ctx context.Context, tx *sql.Tx, profileId uuid.UUID, education models.UniversityEducation) error {
	var universityID, facultyID pgtype.Int4

	err := tx.QueryRowContext(ctx, InsertOrGetUniversityQuery,
		education.University,
		education.City,
	).Scan(&universityID)
	if err != nil {
		return fmt.Errorf("unable to insert/get university: %w", err)
	}

	err = tx.QueryRowContext(ctx, InsertOrGetFacultyQuery, universityID, education.Faculty).Scan(&facultyID)
	if err != nil {
		return fmt.Errorf("unable to insert/get faculty: %w", err)
	}

	_, err = tx.ExecContext(ctx, InsertOrUpdateEducationQuery,
		profileId, facultyID, education.GraduationYear,
	)
	if err != nil {
		return fmt.Errorf("unable to insert/update education: %w", err)
	}

	return nil
}
