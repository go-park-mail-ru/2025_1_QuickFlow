package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
)

const (
	insertChatQuery = `
        INSERT INTO chat (id, name, avatar_url, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
`
	getUserChatsQuery = `
        SELECT c.id, c.name, c.avatar_url, c.type, c.created_at, c.updated_at
        FROM chat c
        join chat_user cu on c.id = cu.chat_id
        WHERE cu.user_id = $1
`
)

type ChatRepository struct {
	connPool *pgxpool.Pool
}

func (c *ChatRepository) CreateChat(ctx context.Context, chat models.Chat) error {
	switch chat.Type {
	case models.ChatTypePrivate:
		_, err := c.connPool.Exec(ctx, insertChatQuery, chat.ID, nil, nil, chat.Type, chat.CreatedAt, chat.UpdatedAt)
		if err != nil {
			return err
		}
	case models.ChatTypeGroup:
		_, err := c.connPool.Exec(ctx, insertChatQuery, chat.ID, chat.Name, chat.AvatarURL, chat.Type, chat.CreatedAt, chat.UpdatedAt)
		if err != nil {
			return err
		}
	default:
		return usecase.ErrInvalidChatType
	}

	return nil
}

func (c *ChatRepository) GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error) {
	var chats []models.Chat
	rows, err := c.connPool.Query(ctx, getUserChatsQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatPostgres pgmodels.ChatPostgres

	for rows.Next() {
		err = rows.Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL, &chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chats = append(chats, *chatPostgres.ToChat())
	}

	return chats, nil
}

//GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error)
//Exists(ctx context.Context, chatId uuid.UUID) (bool, error)
//DeleteChat(ctx context.Context, chatId uuid.UUID) error
//IsParticipant(ctx context.Context, chatId, userId uuid.UUID) (bool, error)
//JoinChat(ctx context.Context, chatId, userId uuid.UUID) error
//LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error
