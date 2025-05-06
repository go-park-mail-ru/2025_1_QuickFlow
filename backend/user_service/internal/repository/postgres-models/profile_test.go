package postgres_models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestConvertToProfile(t *testing.T) {
	tests := []struct {
		name            string
		profilePostgres ProfilePostgres
		expectedProfile models.Profile
	}{
		{
			name: "Success - All fields populated",
			profilePostgres: ProfilePostgres{
				Id:            pgtype.UUID{Bytes: uuid.MustParse("f626ddb6-e69a-42a4-875c-c0b87ce3d4c2")},
				Name:          pgtype.Text{String: "John", Valid: true},
				Surname:       pgtype.Text{String: "Doe", Valid: true},
				Sex:           pgtype.Int4{Int32: 1, Valid: true},
				DateOfBirth:   pgtype.Date{Time: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				Bio:           pgtype.Text{String: "Sample Bio", Valid: true},
				AvatarUrl:     pgtype.Text{String: "avatarUrl", Valid: true},
				BackgroundUrl: pgtype.Text{String: "backgroundUrl", Valid: true},
				LastSeen:      pgtype.Timestamptz{Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				ContactInfo: &ContactInfoPostgres{
					City:  pgtype.Text{String: "New York", Valid: true},
					Email: pgtype.Text{String: "john.doe@example.com", Valid: true},
					Phone: pgtype.Text{String: "1234567890", Valid: true},
				},
				SchoolEducation: &SchoolEducation{
					City:   pgtype.Text{String: "NYC", Valid: true},
					School: pgtype.Text{String: "ABC High School", Valid: true},
				},
				UniversityEducation: &UniversityEducation{
					City:           pgtype.Text{String: "NYC", Valid: true},
					University:     pgtype.Text{String: "XYZ University", Valid: true},
					Faculty:        pgtype.Text{String: "Engineering", Valid: true},
					GraduationYear: pgtype.Int4{Int32: 2025, Valid: true},
				},
			},
			expectedProfile: models.Profile{
				UserId: [16]byte{0xf6, 0x26, 0xdd, 0xb6, 0xe6, 0x9a, 0x42, 0xa4, 0x87, 0x5c, 0xc0, 0xb8, 0x7c, 0xe3, 0xd4, 0xc2},
				BasicInfo: &models.BasicInfo{
					Name:          "John",
					Surname:       "Doe",
					Sex:           models.Sex(1),
					DateOfBirth:   time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
					Bio:           "Sample Bio",
					AvatarUrl:     "avatarUrl",
					BackgroundUrl: "backgroundUrl",
				},
				ContactInfo: &models.ContactInfo{
					City:  "New York",
					Email: "john.doe@example.com",
					Phone: "1234567890",
				},
				SchoolEducation: &models.SchoolEducation{
					City:   "NYC",
					School: "ABC High School",
				},
				UniversityEducation: &models.UniversityEducation{
					City:           "NYC",
					University:     "XYZ University",
					Faculty:        "Engineering",
					GraduationYear: 2025,
				},
				LastSeen: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Error - Missing Contact Info",
			profilePostgres: ProfilePostgres{
				Id:            pgtype.UUID{Bytes: uuid.MustParse("5f626ddb-69a4-42a4-875c-c0b87ce3d4c3")},
				Name:          pgtype.Text{String: "Jane", Valid: true},
				Surname:       pgtype.Text{String: "Doe", Valid: true},
				Sex:           pgtype.Int4{Int32: 2, Valid: true},
				DateOfBirth:   pgtype.Date{Time: time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				Bio:           pgtype.Text{String: "Sample Bio", Valid: true},
				AvatarUrl:     pgtype.Text{String: "avatarUrl", Valid: true},
				BackgroundUrl: pgtype.Text{String: "backgroundUrl", Valid: true},
				LastSeen:      pgtype.Timestamptz{Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			},
			expectedProfile: models.Profile{
				UserId: [16]byte{0x5f, 0x62, 0x6d, 0xdb, 0x69, 0xa4, 0x42, 0xa4, 0x87, 0x5c, 0xc0, 0xb8, 0x7c, 0xe3, 0xd4, 0xc3},
				BasicInfo: &models.BasicInfo{
					Name:          "Jane",
					Surname:       "Doe",
					Sex:           models.Sex(2),
					DateOfBirth:   time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC),
					Bio:           "Sample Bio",
					AvatarUrl:     "avatarUrl",
					BackgroundUrl: "backgroundUrl",
				},
				ContactInfo:         nil,
				SchoolEducation:     nil,
				UniversityEducation: nil,
				LastSeen:            time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := tt.profilePostgres.ConvertToProfile()
			assert.Equal(t, tt.expectedProfile, profile)
		})
	}
}

func TestConvertToPublicUserInfo(t *testing.T) {
	tests := []struct {
		name                   string
		publicUserInfoPostgres PublicUserInfoPostgres
		expectedPublicUserInfo models.PublicUserInfo
	}{
		{
			name: "Success - All fields populated",
			publicUserInfoPostgres: PublicUserInfoPostgres{
				Id:        pgtype.UUID{Bytes: uuid.MustParse("f626ddb6-e69a-42a4-875c-c0b87ce3d4c2")},
				Username:  pgtype.Text{String: "john_doe", Valid: true},
				Firstname: pgtype.Text{String: "John", Valid: true},
				Lastname:  pgtype.Text{String: "Doe", Valid: true},
				AvatarURL: pgtype.Text{String: "avatarUrl", Valid: true},
				LastSeen:  pgtype.Timestamptz{Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			},
			expectedPublicUserInfo: models.PublicUserInfo{
				Id:        [16]byte{0xf6, 0x26, 0xdd, 0xb6, 0xe6, 0x9a, 0x42, 0xa4, 0x87, 0x5c, 0xc0, 0xb8, 0x7c, 0xe3, 0xd4, 0xc2},
				Username:  "john_doe",
				Firstname: "John",
				Lastname:  "Doe",
				AvatarURL: "avatarUrl",
				LastSeen:  time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publicUserInfo := tt.publicUserInfoPostgres.ConvertToPublicUserInfo()
			assert.Equal(t, tt.expectedPublicUserInfo, publicUserInfo)
		})
	}
}
