package forms

import (
	"quickflow/monolith/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProfileForm_FormToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   ProfileForm
		want    models.Profile
		wantErr bool
	}{
		{
			name: "Valid ProfileForm with all fields",
			input: ProfileForm{
				Id: "123e4567-e89b-12d3-a456-426614174000", // Пример правильного UUID
				ProfileInfo: &ProfileInfo{
					Username:    "testuser",
					Name:        "John",
					Surname:     "Doe",
					Sex:         models.MALE,
					DateOfBirth: "1990-01-01",
					Bio:         "Test bio",
				},
				ContactInfo: &ContactInfo{
					City:  "CityName",
					Email: "test@example.com",
					Phone: "123456789",
				},
				SchoolEducation: &SchoolEducationForm{
					SchoolCity: "SchoolCity",
					SchoolName: "TestSchool",
				},
				UniversityEducation: &UniversityEducationForm{
					UniversityCity:    "UniCity",
					UniversityName:    "TestUniversity",
					UniversityFaculty: "FacultyName",
					GraduationYear:    2025,
				},
			},
			want: models.Profile{
				UserId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				BasicInfo: &models.BasicInfo{
					Name:        "John",
					Surname:     "Doe",
					Sex:         models.MALE,
					DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					Bio:         "Test bio",
				},
				ContactInfo: &models.ContactInfo{
					City:  "CityName",
					Email: "test@example.com",
					Phone: "123456789",
				},
				SchoolEducation: &models.SchoolEducation{
					City:   "SchoolCity",
					School: "TestSchool",
				},
				UniversityEducation: &models.UniversityEducation{
					City:           "UniCity",
					University:     "TestUniversity",
					Faculty:        "FacultyName",
					GraduationYear: 2025,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Date format in ProfileInfo",
			input: ProfileForm{
				ProfileInfo: &ProfileInfo{
					Username:    "testuser",
					Name:        "John",
					Surname:     "Doe",
					Sex:         models.MALE,
					DateOfBirth: "invalid-date", // Некорректный формат даты
					Bio:         "Test bio",
				},
			},
			want:    models.Profile{},
			wantErr: true,
		},
		{
			name: "Empty ProfileForm",
			input: ProfileForm{
				Id: "123e4567-e89b-12d3-a456-426614174000",
			},
			want: models.Profile{
				UserId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.FormToModel()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProfileForm.FormToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModelToForm(t *testing.T) {
	var bl = true
	tests := []struct {
		name     string
		input    models.Profile
		isOnline bool
		relation models.UserRelation
		chatId   *uuid.UUID
		want     ProfileForm
	}{
		{
			name: "Valid Profile to ProfileForm",
			input: models.Profile{
				UserId: uuid.MustParse("5b0cc8ec-ee84-4974-8381-612d18b9c6e3"), // Используем фиксированный UUID
				BasicInfo: &models.BasicInfo{
					Name:        "John",
					Surname:     "Doe",
					Sex:         models.MALE,
					DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					Bio:         "Test bio",
				},
				ContactInfo: &models.ContactInfo{
					City:  "CityName",
					Email: "test@example.com",
					Phone: "123456789",
				},
				SchoolEducation: &models.SchoolEducation{
					City:   "SchoolCity",
					School: "TestSchool",
				},
				UniversityEducation: &models.UniversityEducation{
					City:           "UniCity",
					University:     "TestUniversity",
					Faculty:        "FacultyName",
					GraduationYear: 2025,
				},
				LastSeen: time.Now(),
			},
			isOnline: true,
			relation: models.RelationNone,
			chatId:   nil,
			want: ProfileForm{
				Id: "5b0cc8ec-ee84-4974-8381-612d18b9c6e3", // Используем фиксированный UUID
				ProfileInfo: &ProfileInfo{
					Username:    "testuser",
					Name:        "John",
					Surname:     "Doe",
					Sex:         models.MALE,
					DateOfBirth: "1990-01-01",
					Bio:         "Test bio",
				},
				ContactInfo: &ContactInfo{
					City:  "CityName",
					Email: "test@example.com",
					Phone: "123456789",
				},
				SchoolEducation: &SchoolEducationForm{
					SchoolCity: "SchoolCity",
					SchoolName: "TestSchool",
				},
				UniversityEducation: &UniversityEducationForm{
					UniversityCity:    "UniCity",
					UniversityName:    "TestUniversity",
					UniversityFaculty: "FacultyName",
					GraduationYear:    2025,
				},
				IsOnline: &bl,
				Relation: models.RelationNone,
				ChatId:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelToForm(tt.input, "testuser", tt.isOnline, tt.relation, tt.chatId)
			assert.Equal(t, tt.want, got)
		})
	}
}
