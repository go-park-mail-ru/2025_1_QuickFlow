package postgres

import (
	"context"
	"fmt"
	models2 "quickflow/monolith/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestSaveProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile models2.Profile
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successful save profile",
			profile: models2.Profile{
				UserId: uuid.New(),
				BasicInfo: &models2.BasicInfo{
					Bio:           "Test bio",
					AvatarUrl:     "http://avatar.url",
					BackgroundUrl: "http://background.url",
					Name:          "John",
					Surname:       "Doe",
					Sex:           models2.MALE,
					DateOfBirth:   time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into profile`).WithArgs(
					sqlmock.AnyArg(), "Test bio", "http://avatar.url", "http://background.url",
					"John", "Doe", models2.MALE, time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
				).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Failed to save profile",
			profile: models2.Profile{
				UserId: uuid.New(),
				BasicInfo: &models2.BasicInfo{
					Bio:           "Test bio",
					AvatarUrl:     "http://avatar.url",
					BackgroundUrl: "http://background.url",
					Name:          "John",
					Surname:       "Doe",
					Sex:           models2.MALE,
					DateOfBirth:   time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into profile`).WithArgs(
					sqlmock.AnyArg(), "Test bio", "http://avatar.url", "http://background.url",
					"John", "Doe", models2.MALE, time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
				).WillReturnError(fmt.Errorf("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			pgRepo := &PostgresProfileRepository{connPool: mockDB}
			tt.mock(mock)

			err = pgRepo.SaveProfile(context.Background(), tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveProfile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name    string
		userId  uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    models2.Profile
		wantErr bool
	}{
		{
			name:   "Successfully get profile",
			userId: uuid_,
			mock: func(mock sqlmock.Sqlmock) {
				// Мокаем запрос для получения профиля
				mock.ExpectQuery(`select id, bio, profile_avatar`).WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "bio", "profile_avatar", "profile_background", "firstname", "lastname", "sex", "birth_date", "school_id", "contact_info_id", "last_seen"}).
						AddRow(uuid_, "Test bio", "http://avatar.url", "http://background.url", "John", "Doe", models2.MALE, time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
							pgtype.UUID{Valid: false}, pgtype.UUID{Valid: false}, time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)))

				// Добавляем мок для запроса получения образования
				mock.ExpectQuery(`select u.name, u.city, f.name, e.graduation_year`).WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"name", "city", "faculty", "graduation_year"}).
						AddRow("University Name", "City", "Faculty Name", 2021))
			},
			want: models2.Profile{
				UserId: uuid_,
				BasicInfo: &models2.BasicInfo{
					Bio:           "Test bio",
					AvatarUrl:     "http://avatar.url",
					BackgroundUrl: "http://background.url",
					Name:          "John",
					Surname:       "Doe",
					Sex:           models2.MALE,
					DateOfBirth:   time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
				UniversityEducation: &models2.UniversityEducation{
					University:     "University Name",
					City:           "City",
					Faculty:        "Faculty Name",
					GraduationYear: 2021,
				},
				LastSeen: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:   "Failed to get profile",
			userId: uuid.New(),
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, bio, profile_avatar`).WithArgs(sqlmock.AnyArg()).
					WillReturnError(fmt.Errorf("select failed"))
			},
			want:    models2.Profile{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			pgRepo := &PostgresProfileRepository{connPool: mockDB}
			tt.mock(mock)

			got, err := pgRepo.GetProfile(context.Background(), tt.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProfile() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateProfileTextInfo(t *testing.T) {
	tests := []struct {
		name    string
		profile models2.Profile
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successfully update profile",
			profile: models2.Profile{
				UserId: uuid.New(),
				BasicInfo: &models2.BasicInfo{
					Bio:           "Updated bio",
					AvatarUrl:     "http://updated-avatar.url",
					BackgroundUrl: "http://updated-background.url",
					Name:          "Updated John",
					Surname:       "Updated Doe",
					Sex:           models2.MALE,
					DateOfBirth:   time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`update profile set bio`).WithArgs(
					sqlmock.AnyArg(), "Updated bio", "Updated John", "Updated Doe", models2.MALE, time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
				).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Failed to update profile",
			profile: models2.Profile{
				UserId: uuid.New(),
				BasicInfo: &models2.BasicInfo{
					Bio:           "Updated bio",
					AvatarUrl:     "http://updated-avatar.url",
					BackgroundUrl: "http://updated-background.url",
					Name:          "Updated John",
					Surname:       "Updated Doe",
					Sex:           models2.MALE,
					DateOfBirth:   time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`update profile set bio`).WithArgs(
					sqlmock.AnyArg(), "Updated bio", "Updated John", "Updated Doe", models2.MALE, time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
				).WillReturnError(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			pgRepo := &PostgresProfileRepository{connPool: mockDB}
			tt.mock(mock)

			err = pgRepo.UpdateProfileTextInfo(context.Background(), tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProfileTextInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
