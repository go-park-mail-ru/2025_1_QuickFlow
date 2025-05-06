package postgres_models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

func TestUserPostgres_ConvertToUser(t *testing.T) {
	testUUID := uuid.New()

	tests := []struct {
		name     string
		input    UserPostgres
		expected models.User
	}{
		{
			name: "All fields valid",
			input: UserPostgres{
				Id:       pgtype.UUID{Bytes: testUUID, Valid: true},
				Username: pgtype.Text{String: "testuser", Valid: true},
				Password: pgtype.Text{String: "hashedpass", Valid: true},
				Salt:     pgtype.Text{String: "somesalt", Valid: true},
			},
			expected: models.User{
				Id:       testUUID,
				Username: "testuser",
				Password: "hashedpass",
				Salt:     "somesalt",
			},
		},
		{
			name: "Some fields invalid",
			input: UserPostgres{
				Id:       pgtype.UUID{Bytes: testUUID, Valid: true},
				Username: pgtype.Text{Valid: false},
				Password: pgtype.Text{String: "hashedpass", Valid: true},
				Salt:     pgtype.Text{Valid: false},
			},
			expected: models.User{
				Id:       testUUID,
				Username: "",
				Password: "hashedpass",
				Salt:     "",
			},
		},
		{
			name: "All fields invalid",
			input: UserPostgres{
				Id:       pgtype.UUID{Valid: false},
				Username: pgtype.Text{Valid: false},
				Password: pgtype.Text{Valid: false},
				Salt:     pgtype.Text{Valid: false},
			},
			expected: models.User{
				Id:       uuid.Nil,
				Username: "",
				Password: "",
				Salt:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ConvertToUser()

			if result.Id != tt.expected.Id {
				t.Errorf("Id: got %v, want %v", result.Id, tt.expected.Id)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Username: got %v, want %v", result.Username, tt.expected.Username)
			}
			if result.Password != tt.expected.Password {
				t.Errorf("Password: got %v, want %v", result.Password, tt.expected.Password)
			}
			if result.Salt != tt.expected.Salt {
				t.Errorf("Salt: got %v, want %v", result.Salt, tt.expected.Salt)
			}
		})
	}
}

func TestConvertUserToPostgres(t *testing.T) {
	testUUID := uuid.New()

	tests := []struct {
		name     string
		input    models.User
		expected UserPostgres
	}{
		{
			name: "All fields populated",
			input: models.User{
				Id:       testUUID,
				Username: "testuser",
				Password: "hashedpass",
				Salt:     "somesalt",
			},
			expected: UserPostgres{
				Id:       pgtype.UUID{Bytes: testUUID, Valid: true},
				Username: pgtype.Text{String: "testuser", Valid: true},
				Password: pgtype.Text{String: "hashedpass", Valid: true},
				Salt:     pgtype.Text{String: "somesalt", Valid: true},
			},
		},
		{
			name: "Empty fields",
			input: models.User{
				Id:       uuid.Nil,
				Username: "",
				Password: "",
				Salt:     "",
			},
			expected: UserPostgres{
				Id:       pgtype.UUID{Bytes: uuid.Nil, Valid: true},
				Username: pgtype.Text{String: "", Valid: true},
				Password: pgtype.Text{String: "", Valid: true},
				Salt:     pgtype.Text{String: "", Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertUserToPostgres(tt.input)

			if result.Id.Bytes != tt.expected.Id.Bytes || result.Id.Valid != tt.expected.Id.Valid {
				t.Errorf("Id: got %v (valid: %t), want %v (valid: %t)",
					result.Id.Bytes, result.Id.Valid, tt.expected.Id.Bytes, tt.expected.Id.Valid)
			}
			if result.Username.String != tt.expected.Username.String || result.Username.Valid != tt.expected.Username.Valid {
				t.Errorf("Username: got %v (valid: %t), want %v (valid: %t)",
					result.Username.String, result.Username.Valid, tt.expected.Username.String, tt.expected.Username.Valid)
			}
			if result.Password.String != tt.expected.Password.String || result.Password.Valid != tt.expected.Password.Valid {
				t.Errorf("Password: got %v (valid: %t), want %v (valid: %t)",
					result.Password.String, result.Password.Valid, tt.expected.Password.String, tt.expected.Password.Valid)
			}
			if result.Salt.String != tt.expected.Salt.String || result.Salt.Valid != tt.expected.Salt.Valid {
				t.Errorf("Salt: got %v (valid: %t), want %v (valid: %t)",
					result.Salt.String, result.Salt.Valid, tt.expected.Salt.String, tt.expected.Salt.Valid)
			}
		})
	}
}

func TestConvertStringToPostgresText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected pgtype.Text
	}{
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: pgtype.Text{String: "hello", Valid: true},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: pgtype.Text{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStringToPostgresText(tt.input)

			if result.String != tt.expected.String || result.Valid != tt.expected.Valid {
				t.Errorf("got %v (valid: %t), want %v (valid: %t)",
					result.String, result.Valid, tt.expected.String, tt.expected.Valid)
			}
		})
	}
}
