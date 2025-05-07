package postgres_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/shared/models"
	"testing"
	"time"
)

func TestCreateChat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		chat      models.Chat
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success create private chat",
			chat: models.Chat{
				ID:        uuid.New(),
				Type:      models.ChatTypePrivate,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO chat`).
					WithArgs(sqlmock.AnyArg(), nil, nil, models.ChatTypePrivate, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "success create group chat",
			chat: models.Chat{
				ID:        uuid.New(),
				Name:      "Study Group",
				AvatarURL: "https://img",
				Type:      models.ChatTypeGroup,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO chat`).
					WithArgs(sqlmock.AnyArg(), "Study Group", "https://img", models.ChatTypeGroup, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "invalid chat type",
			chat: models.Chat{
				ID:        uuid.New(),
				Type:      99, // невалидный
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// не вызывается вообще
			},
			wantErr: true,
		},
		{
			name: "db error on private chat",
			chat: models.Chat{
				ID:        uuid.New(),
				Type:      models.ChatTypePrivate,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO chat`).
					WithArgs(sqlmock.AnyArg(), nil, nil, models.ChatTypePrivate, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создание нового mock подключения
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			repo := postgres.NewPostgresChatRepository(db)

			err = repo.CreateChat(ctx, tt.chat)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Проверяем, что все ожидания были выполнены
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
