package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"quickflow/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"quickflow/config/postgres"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
)

const (
	getMessagesForChatOlderQuery = `
        SELECT id, chat_id, sender_id, text, created_at, updated_at, is_read
        FROM message
        WHERE chat_id = $1 AND created_at < $2
        ORDER BY created_at DESC
        LIMIT $3
    `

	getFilesQuery = `
        SELECT file_url
        FROM message_file
        WHERE message_id = $1
`
	saveMessageQuery = `
        INSERT INTO message (id, chat_id, sender_id, text, created_at, updated_at, is_read)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	markReadQuery = `
        UPDATE message
        SET is_read = true
        WHERE id = $1
`

	getLastChatMessage = `
    with otv as (
        select * from message m 
        where m.chat_id = $1
    )
    select c.id, c.chat_id, c.sender_id, c.text, c.created_at, c.updated_at, c.is_read
    from (select * from otv) c
    where c.created_at = (
        select max(created_at) 
        from otv);
`
)

type MessageRepository struct {
	connPool *pgxpool.Pool
}

func NewPostgresMessageRepository() *MessageRepository {
	connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &MessageRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (m *MessageRepository) Close() {
	m.connPool.Close()
}

func (m *MessageRepository) GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Message, error) {
	rows, err := m.connPool.Query(ctx, getMessagesForChatOlderQuery, chatId, timestamp, numPosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var messagePostgres pgmodels.MessagePostgres
		if err := rows.Scan(&messagePostgres.ID, &messagePostgres.ChatID, &messagePostgres.SenderID,
			&messagePostgres.Text, &messagePostgres.CreatedAt, &messagePostgres.UpdatedAt, &messagePostgres.IsRead); err != nil {
			return nil, err
		}

		message := messagePostgres.ToMessage()
		files, err := m.connPool.Query(ctx, getFilesQuery, messagePostgres.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		for files.Next() {
			var fileURL pgtype.Text
			err = files.Scan(&fileURL)
			if err != nil {
				return nil, err
			}
			if fileURL.Valid {
				message.AttachmentURLs = append(message.AttachmentURLs, fileURL.String)
			}
		}
		files.Close()

		messages = append(messages, message)
	}

	return messages, nil
}

func (m *MessageRepository) SaveMessage(ctx context.Context, message models.Message) error {
	messagePostgres := pgmodels.FromMessage(message)
	_, err := m.connPool.Exec(ctx, saveMessageQuery,
		messagePostgres.ID, messagePostgres.ChatID, messagePostgres.SenderID,
		messagePostgres.Text, messagePostgres.CreatedAt, messagePostgres.UpdatedAt,
		messagePostgres.IsRead)
	if err != nil {
		return fmt.Errorf("unable to save message to database: %w", err)
	}
	for _, fileURL := range messagePostgres.AttachmentsURLs {
		_, err = m.connPool.Exec(ctx, "INSERT INTO message_file (message_id, file_url) VALUES ($1, $2)",
			messagePostgres.ID, fileURL)
		if err != nil {
			return fmt.Errorf("unable to save file URL to database: %w", err)
		}
	}

	_, err = m.connPool.Exec(ctx, `update chat set updated_at = $1 where id = $2`,
		messagePostgres.UpdatedAt, messagePostgres.ChatID)
	if err != nil {
		logger.Error(ctx, "Unable to update chat updated_at: ", err)
		return fmt.Errorf("unable to update chat updated_at: %w", err)
	}
	return nil
}

func (m *MessageRepository) DeleteMessage(ctx context.Context, messageId uuid.UUID) error {
	_, err := m.connPool.Exec(ctx, "DELETE FROM message WHERE id = $1", messageId)
	if err != nil {
		return fmt.Errorf("unable to delete message from database: %w", err)
	}
	return nil
}
func (m *MessageRepository) MarkRead(ctx context.Context, messageId uuid.UUID) error {
	_, err := m.connPool.Exec(ctx, markReadQuery, messageId)
	if err != nil {
		return fmt.Errorf("unable to mark message as read: %w", err)
	}
	return nil
}

func (m *MessageRepository) GetLastChatMessage(ctx context.Context, chatId uuid.UUID) (*models.Message, error) {
	var messagePostgres pgmodels.MessagePostgres
	err := m.connPool.QueryRow(ctx, getLastChatMessage, chatId).Scan(
		&messagePostgres.ID, &messagePostgres.ChatID, &messagePostgres.SenderID,
		&messagePostgres.Text, &messagePostgres.CreatedAt, &messagePostgres.UpdatedAt,
		&messagePostgres.IsRead)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		logger.Error(ctx, "Unable to get last message from database: ", err)
		return nil, fmt.Errorf("unable to get last message from database: %w", err)
	}

	message := messagePostgres.ToMessage()
	return &message, nil
}
