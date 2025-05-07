package postgres_models

import (
	"testing"
	"time"

	"github.com/google/uuid" // Для работы с UUID
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"quickflow/shared/models"
)

func TestMessagePostgres_ToMessage(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name     string
		input    *MessagePostgres
		expected *models.Message
	}{
		{
			name: "Successfully convert MessagePostgres to Message",
			input: &MessagePostgres{
				ID:              pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:            pgtype.Text{String: "Test message", Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AttachmentsURLs: []pgtype.Text{{String: "http://example.com/file1.png", Valid: true}},
				SenderID:        pgtype.UUID{Bytes: uuid_, Valid: true},
				ChatID:          pgtype.UUID{Bytes: uuid_, Valid: true},
			},
			expected: &models.Message{
				ID:             uuid_, // UUID генерируется на лету
				Text:           "Test message",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				AttachmentURLs: []string{"http://example.com/file1.png"},
				SenderID:       uuid_, // UUID генерируется на лету
				ChatID:         uuid_, // UUID генерируется на лету
			},
		},
		{
			name: "Message without AttachmentsURLs",
			input: &MessagePostgres{
				ID:              pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:            pgtype.Text{String: "Message without attachments", Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AttachmentsURLs: []pgtype.Text{},
				SenderID:        pgtype.UUID{Bytes: uuid_, Valid: true},
				ChatID:          pgtype.UUID{Bytes: uuid_, Valid: true},
			},
			expected: &models.Message{
				ID:             uuid_, // UUID генерируется на лету
				Text:           "Message without attachments",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				AttachmentURLs: []string{},
				SenderID:       uuid_, // UUID генерируется на лету
				ChatID:         uuid_, // UUID генерируется на лету
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ToMessage()

			// Для проверки времени можно использовать подход с округлением до секунд
			assert.Equal(t, tt.expected.ID, got.ID)
			assert.Equal(t, tt.expected.Text, got.Text)
			assert.WithinDuration(t, tt.expected.CreatedAt, got.CreatedAt, time.Second)
			assert.WithinDuration(t, tt.expected.UpdatedAt, got.UpdatedAt, time.Second)
			assert.ElementsMatch(t, tt.expected.AttachmentURLs, got.AttachmentURLs)
			assert.Equal(t, tt.expected.SenderID, got.SenderID)
			assert.Equal(t, tt.expected.ChatID, got.ChatID)
		})
	}
}

func TestFromMessage(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name     string
		input    *models.Message
		expected *MessagePostgres
	}{
		{
			name: "Successfully convert Message to MessagePostgres",
			input: &models.Message{
				ID:             uuid_, // Генерация UUID
				Text:           "Test message",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				AttachmentURLs: []string{"http://example.com/file1.png"},
				SenderID:       uuid_, // Генерация UUID
				ChatID:         uuid_, // Генерация UUID
			},
			expected: &MessagePostgres{
				ID:              pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:            pgtype.Text{String: "Test message", Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AttachmentsURLs: []pgtype.Text{{String: "http://example.com/file1.png", Valid: true}},
				SenderID:        pgtype.UUID{Bytes: uuid_, Valid: true},
				ChatID:          pgtype.UUID{Bytes: uuid_, Valid: true},
			},
		},
		{
			name: "Message without AttachmentsURLs",
			input: &models.Message{
				ID:             uuid_, // Генерация UUID
				Text:           "Message without attachments",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				AttachmentURLs: []string{},
				SenderID:       uuid_, // Генерация UUID
				ChatID:         uuid_, // Генерация UUID
			},
			expected: &MessagePostgres{
				ID:              pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:            pgtype.Text{String: "Message without attachments", Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AttachmentsURLs: []pgtype.Text{},
				SenderID:        pgtype.UUID{Bytes: uuid_, Valid: true},
				ChatID:          pgtype.UUID{Bytes: uuid_, Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromMessage(*tt.input)

			// Для проверки времени можно использовать подход с округлением до секунд
			assert.Equal(t, tt.expected.ID.Bytes, got.ID.Bytes)
			assert.Equal(t, tt.expected.Text, got.Text)
			assert.WithinDuration(t, tt.expected.CreatedAt.Time, got.CreatedAt.Time, time.Second)
			assert.WithinDuration(t, tt.expected.UpdatedAt.Time, got.UpdatedAt.Time, time.Second)
			assert.ElementsMatch(t, tt.expected.AttachmentsURLs, got.AttachmentsURLs)
			assert.Equal(t, tt.expected.SenderID.Bytes, got.SenderID.Bytes)
			assert.Equal(t, tt.expected.ChatID.Bytes, got.ChatID.Bytes)
		})
	}
}
