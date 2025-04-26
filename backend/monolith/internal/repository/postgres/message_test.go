package postgres_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"quickflow/internal/models"
	"quickflow/internal/repository/postgres"
	postgresmodels "quickflow/internal/repository/postgres/postgres-models"
	"testing"
	"time"
)

func TestMessageRepository(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		message   models.Message
		mockSetup func(mock sqlmock.Sqlmock, msg models.Message)
		wantErr   bool
	}{
		{
			name:    "success save message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				pgMsg := postgresmodels.FromMessage(msg)
				mock.ExpectExec(`(?i)INSERT INTO message`).
					WithArgs(
						pgMsg.ID,
						pgMsg.ChatID,
						pgMsg.SenderID,
						pgMsg.Text,
						pgMsg.CreatedAt,
						pgMsg.UpdatedAt,
						pgMsg.ReadAt,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(`(?i)update chat set updated_at = \$1 where id = \$2`).
					WithArgs(pgMsg.UpdatedAt, pgMsg.ChatID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:    "db error on save message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				pgMsg := postgresmodels.FromMessage(msg)
				mock.ExpectExec(`(?i)INSERT INTO message`).
					WithArgs(pgMsg.ID, pgMsg.ChatID, pgMsg.SenderID, pgMsg.Text, pgMsg.CreatedAt, pgMsg.UpdatedAt, pgMsg.ReadAt).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:    "success mark message as read",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				mock.ExpectExec(`(?i)UPDATE message`).
					WithArgs(msg.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:    "db error on mark message as read",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				mock.ExpectExec(`(?i)UPDATE message`).
					WithArgs(msg.ID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:    "success get messages for chat",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				pgMsg := postgresmodels.FromMessage(msg)
				mock.ExpectQuery(`(?i)SELECT id, chat_id, sender_id, text, created_at, updated_at, is_read`).
					WithArgs(pgMsg.ChatID, sqlmock.AnyArg(), 10). // timestamp как anyArg
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "chat_id", "sender_id", "text", "created_at", "updated_at", "is_read",
					}).AddRow(pgMsg.ID, pgMsg.ChatID, pgMsg.SenderID, pgMsg.Text, pgMsg.CreatedAt, pgMsg.UpdatedAt, pgMsg.ReadAt))

				mock.ExpectQuery(`(?i)SELECT file_url`).
					WithArgs(pgMsg.ID). // Ищем файлы по ID сообщения
					WillReturnRows(sqlmock.NewRows([]string{"file_url"}).
						AddRow("file:///path/to/file.txt")) // Мокируем путь к файлу
			},
			wantErr: false,
		},
		{
			name:    "db error on get messages for chat",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				pgMsg := postgresmodels.FromMessage(msg)
				mock.ExpectQuery(`(?i)SELECT id, chat_id, sender_id, text, created_at, updated_at, is_read`).
					WithArgs(pgMsg.ChatID, sqlmock.AnyArg(), 10).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:    "success delete message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				mock.ExpectExec(`(?i)DELETE FROM message`).
					WithArgs(msg.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:    "db error on delete message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				mock.ExpectExec(`(?i)DELETE FROM message`).
					WithArgs(msg.ID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},

		// тест для GetLastChatMessage
		{
			name:    "success get last chat message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				pgMsg := postgresmodels.FromMessage(msg)
				mock.ExpectQuery(`(?i)WITH otv AS`).
					WithArgs(msg.ChatID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "chat_id", "sender_id", "text", "created_at", "updated_at", "is_read",
					}).AddRow(pgMsg.ID, pgMsg.ChatID, pgMsg.SenderID, pgMsg.Text, pgMsg.CreatedAt, pgMsg.UpdatedAt, pgMsg.ReadAt))
			},
			wantErr: false,
		},
		{
			name:    "db error on get last chat message",
			message: newTestMessage(),
			mockSetup: func(mock sqlmock.Sqlmock, msg models.Message) {
				mock.ExpectQuery(`(?i)WITH otv AS`).
					WithArgs(msg.ChatID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New() // Создание mock-объекта для *sql.DB
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.message)
			}

			repo := postgres.NewPostgresMessageRepository(db)

			// Здесь тестируем различные случаи на основе тестовых данных
			switch tt.name {
			case "success save message", "db error on save message":
				err = repo.SaveMessage(ctx, tt.message)
			case "success mark message as read", "db error on mark message as read":
				err = repo.UpdateLastMessageRead(ctx, tt.message.ID)
			case "success get messages for chat", "db error on get messages for chat":
				_, err = repo.GetMessagesForChatOlder(ctx, tt.message.ChatID, 10, time.Now())
			case "success get last chat message", "db error on get last chat message":
				_, err = repo.GetLastChatMessage(ctx, tt.message.ChatID)
			case "success delete message", "db error on delete message":
				err = repo.DeleteMessage(ctx, tt.message.ID)
			}

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet()) // Проверка всех ожиданий
		})
	}
}

func newTestMessage() models.Message {
	return models.Message{
		ID:        uuid.New(),
		ChatID:    uuid.New(),
		SenderID:  uuid.New(),
		Text:      "Hello, World!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsRead:    false,
	}
}
