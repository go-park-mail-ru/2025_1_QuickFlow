package postgres_models

import (
	"testing"
	"time"

	"github.com/google/uuid" // Для работы с UUID
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"quickflow/shared/models"
)

func TestChatPostgres_ToChat(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name     string
		input    *ChatPostgres
		expected *models.Chat
	}{
		{
			name: "Successfully convert to Chat",
			input: &ChatPostgres{
				Id:        pgtype.UUID{Bytes: uuid_, Valid: true},
				Name:      pgtype.Text{String: "Test Chat", Valid: true},
				AvatarURL: pgtype.Text{String: "http://example.com/avatar.png", Valid: true},
				Type:      pgtype.Int4{Int32: 1, Valid: true},
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
			expected: &models.Chat{
				ID:        uuid_, // Строка UUID или определённый UUID
				Name:      "Test Chat",
				AvatarURL: "http://example.com/avatar.png",
				Type:      models.ChatType(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "Chat without AvatarURL and Name",
			input: &ChatPostgres{
				Id:        pgtype.UUID{Bytes: uuid_, Valid: true},
				Name:      pgtype.Text{Valid: false},
				AvatarURL: pgtype.Text{Valid: false},
				Type:      pgtype.Int4{Int32: 2, Valid: true},
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
			expected: &models.Chat{
				ID:        uuid_, // Строка UUID или определённый UUID
				Name:      "",
				AvatarURL: "",
				Type:      models.ChatType(2),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ToChat()

			// Для проверки времени можно использовать подход с округлением до секунд
			assert.Equal(t, tt.expected.ID, got.ID)
			assert.Equal(t, tt.expected.Name, got.Name)
			assert.Equal(t, tt.expected.AvatarURL, got.AvatarURL)
			assert.Equal(t, tt.expected.Type, got.Type)
			assert.WithinDuration(t, tt.expected.CreatedAt, got.CreatedAt, time.Second)
			assert.WithinDuration(t, tt.expected.UpdatedAt, got.UpdatedAt, time.Second)
		})
	}
}

func TestModelToPostgres(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name     string
		input    *models.Chat
		expected *ChatPostgres
	}{
		{
			name: "Successfully convert to ChatPostgres",
			input: &models.Chat{
				ID:        uuid_, // Генерация UUID
				Name:      "Test Chat",
				AvatarURL: "http://example.com/avatar.png",
				Type:      models.ChatType(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: &ChatPostgres{
				Id:        pgtype.UUID{Bytes: uuid_, Valid: true},
				Name:      pgtype.Text{String: "Test Chat", Valid: true},
				AvatarURL: pgtype.Text{String: "http://example.com/avatar.png", Valid: true},
				Type:      pgtype.Int4{Int32: 1, Valid: true},
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		},
		{
			name: "Chat with empty AvatarURL and Name",
			input: &models.Chat{
				ID:        uuid_, // Генерация UUID
				Name:      "",
				AvatarURL: "",
				Type:      models.ChatType(2),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: &ChatPostgres{
				Id:        pgtype.UUID{Bytes: uuid_, Valid: true},
				Name:      pgtype.Text{Valid: false},
				AvatarURL: pgtype.Text{Valid: false},
				Type:      pgtype.Int4{Int32: 2, Valid: true},
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelToPostgres(tt.input)

			// Для проверки времени можно использовать подход с округлением до секунд
			assert.Equal(t, tt.expected.Id.Bytes, got.Id.Bytes)
			assert.Equal(t, tt.expected.Name, got.Name)
			assert.Equal(t, tt.expected.AvatarURL, got.AvatarURL)
			assert.Equal(t, tt.expected.Type, got.Type)
			assert.WithinDuration(t, tt.expected.CreatedAt.Time, got.CreatedAt.Time, time.Second)
			assert.WithinDuration(t, tt.expected.UpdatedAt.Time, got.UpdatedAt.Time, time.Second)
		})
	}
}
