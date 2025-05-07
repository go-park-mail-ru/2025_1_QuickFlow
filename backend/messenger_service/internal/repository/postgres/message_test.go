package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMessageRepository_GetLastChatMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMessageRepository(db)

	chatID := uuid.New()

	// Mocking the last chat message
	rows := sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"}).
		AddRow(uuid.New(), chatID, uuid.New(), "Last message", time.Now(), time.Now())

	mock.ExpectQuery("select c.id, c.chat_id, c.sender_id, c.text, c.created_at, c.updated_at").
		WithArgs(chatID).
		WillReturnRows(rows)

	message, err := repo.GetLastChatMessage(context.Background(), chatID)
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, message.Text, "Last message")
}

func TestMessageRepository_GetMessageById(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMessageRepository(db)

	messageID := uuid.New()

	// Mocking the get message by id
	rows := sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"}).
		AddRow(messageID, uuid.New(), uuid.New(), "Message by ID", time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, chat_id, sender_id, text, created_at, updated_at").
		WithArgs(messageID).
		WillReturnRows(rows)

	message, err := repo.GetMessageById(context.Background(), messageID)
	assert.NoError(t, err)
	assert.Equal(t, message.Text, "Message by ID")
}
