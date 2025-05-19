package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	messenger_service "quickflow/messenger_service/internal/errors"
	pgmodels "quickflow/messenger_service/internal/repository/postgres-models"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

const (
	getMessagesForChatOlderQuery = `
        SELECT id, chat_id, sender_id, text, created_at, updated_at
        FROM message
        WHERE chat_id = $1 AND created_at < $2
        ORDER BY created_at desc 
        LIMIT $3
    `

	getFilesQuery = `
        SELECT file_url, file_type
        FROM message_file
        WHERE message_id = $1
`
	saveFilesQuery = `
	INSERT INTO message_file
	(message_id, file_url, file_type)
	VALUES ($1, $2, $3)
	`

	saveMessageQuery = `
        INSERT INTO message (id, chat_id, sender_id, text, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
`
	markReadQuery = `
        update chat_user
        set last_read = $3
        where chat_id = $1 and user_id = $2;
        
`
	getLastReadMessageQuery = `
		select max(last_read) 
		from chat_user
		where chat_id = $1 and user_id != $2;
`

	getLastChatMessage = `
    with otv as (
        select * from message m 
        where m.chat_id = $1
    )
    select c.id, c.chat_id, c.sender_id, c.text, c.created_at, c.updated_at
    from (select * from otv) c
    where c.created_at = (
        select max(created_at) 
        from otv);
`
)

type MessageRepository struct {
	connPool *sql.DB
}

func NewPostgresMessageRepository(connPool *sql.DB) *MessageRepository {
	return &MessageRepository{
		connPool: connPool,
	}
}

// Close закрывает пул соединений
func (m *MessageRepository) Close() {
	m.connPool.Close()
}

func (m *MessageRepository) GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID,
	numMessages int, timestamp time.Time) ([]models.Message, error) {
	rows, err := m.connPool.QueryContext(ctx, getMessagesForChatOlderQuery, pgtype.UUID{Bytes: chatId, Valid: true},
		pgtype.Timestamptz{Time: timestamp, Valid: true}, numMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var messagePostgres pgmodels.MessagePostgres
		if err := rows.Scan(&messagePostgres.ID, &messagePostgres.ChatID, &messagePostgres.SenderID,
			&messagePostgres.Text, &messagePostgres.CreatedAt, &messagePostgres.UpdatedAt); err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to scan message from database for chat %v, numMessages %v, timestamp %v: %v",
				chatId, numMessages, timestamp, err))
			return nil, err
		}

		message := messagePostgres.ToMessage()
		files, err := m.connPool.QueryContext(ctx, getFilesQuery, messagePostgres.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, fmt.Sprintf("Unable to get files for message %v: %v", messagePostgres.ID, err))
			return nil, err
		}
		for files.Next() {
			var pgfile pgmodels.PostgresFile
			err = files.Scan(&pgfile.URL, &pgfile.DisplayType)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Unable to scan file URL for message %v: %v", messagePostgres.ID, err))
				return nil, err
			}

			message.Attachments = append(message.Attachments, pgfile.ToFile())
		}
		files.Close()

		messages = slices.Insert(messages, 0, message)
	}
	logger.Info(ctx, fmt.Sprintf("Fetched %d messages for chat %s", len(messages), chatId))

	return messages, nil
}

func (m *MessageRepository) SaveMessage(ctx context.Context, message models.Message) error {
	messagePostgres := pgmodels.FromMessage(message)
	_, err := m.connPool.ExecContext(ctx, saveMessageQuery,
		messagePostgres.ID, messagePostgres.ChatID, messagePostgres.SenderID,
		messagePostgres.Text, messagePostgres.CreatedAt, messagePostgres.UpdatedAt)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to save message %v to database: %s", messagePostgres.ID, err.Error()))
		return fmt.Errorf("unable to save message to database: %w", err)
	}
	for _, file := range messagePostgres.Attachments {
		_, err = m.connPool.ExecContext(ctx, saveFilesQuery,
			messagePostgres.ID, file.URL, file.DisplayType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to save file URL %v for message %v to database: %s", file.URL, messagePostgres.ID, err.Error()))
			return fmt.Errorf("unable to save file URL to database: %w", err)
		}
	}

	_, err = m.connPool.ExecContext(ctx, `update chat set updated_at = $1 where id = $2`,
		messagePostgres.UpdatedAt, messagePostgres.ChatID)
	if err != nil {
		logger.Error(ctx, "Unable to update chat updated_at: ", err)
		return fmt.Errorf("unable to update chat updated_at: %w", err)
	}
	return nil
}

func (m *MessageRepository) DeleteMessage(ctx context.Context, messageId uuid.UUID) error {
	_, err := m.connPool.ExecContext(ctx, "DELETE FROM message WHERE id = $1", messageId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to delete message %v from database: %s", messageId, err.Error()))
		return fmt.Errorf("unable to delete message from database: %w", err)
	}
	return nil
}
func (m *MessageRepository) UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId uuid.UUID, userId uuid.UUID) error {
	_, err := m.connPool.ExecContext(ctx, markReadQuery, chatId, userId, pgtype.Timestamptz{Time: timestamp, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		logger.Error(ctx, fmt.Sprintf("Unable to find chat %v with user %v: %s", chatId, userId, err.Error()))
		return messenger_service.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to update last read %v for chat %v with user %v: %s", timestamp, chatId, userId, err.Error()))
		return fmt.Errorf("unable to update last read message in database: %w", err)
	}
	return nil
}

func (m *MessageRepository) GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*time.Time, error) {
	var timestamp pgtype.Timestamptz
	err := m.connPool.QueryRowContext(ctx, getLastReadMessageQuery, pgtype.UUID{Bytes: chatId, Valid: true}, pgtype.UUID{Bytes: userId, Valid: true}).Scan(
		&timestamp)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		logger.Error(ctx, "Unable to get last read message from database: ", err)
		return nil, fmt.Errorf("unable to get last read message from database: %w", err)
	}

	if timestamp.Valid {
		return &timestamp.Time, nil
	}
	return nil, nil
}

func (m *MessageRepository) GetLastChatMessage(ctx context.Context, chatId uuid.UUID) (*models.Message, error) {
	var messagePostgres pgmodels.MessagePostgres
	err := m.connPool.QueryRowContext(ctx, getLastChatMessage, pgtype.UUID{Bytes: chatId, Valid: true}).Scan(
		&messagePostgres.ID, &messagePostgres.ChatID, &messagePostgres.SenderID,
		&messagePostgres.Text, &messagePostgres.CreatedAt, &messagePostgres.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		logger.Error(ctx, "Unable to get last message from database: ", err)
		return nil, fmt.Errorf("unable to get last message from database: %w", err)
	}

	message := messagePostgres.ToMessage()
	return &message, nil
}

func (m *MessageRepository) GetMessageById(ctx context.Context, messageId uuid.UUID) (models.Message, error) {
	var messagePostgres pgmodels.MessagePostgres
	err := m.connPool.QueryRowContext(ctx, "SELECT id, chat_id, sender_id, text, created_at, updated_at FROM message WHERE id = $1", messageId).Scan(
		&messagePostgres.ID, &messagePostgres.ChatID, &messagePostgres.SenderID,
		&messagePostgres.Text, &messagePostgres.CreatedAt, &messagePostgres.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Message{}, messenger_service.ErrNotFound
	} else if err != nil {
		logger.Error(ctx, "Unable to get message from database: ", err)
		return models.Message{}, fmt.Errorf("unable to get message from database: %w", err)
	}

	message := messagePostgres.ToMessage()
	return message, nil
}
