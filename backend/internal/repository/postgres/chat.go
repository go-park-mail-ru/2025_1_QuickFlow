package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"quickflow/config/postgres"
	"quickflow/internal/models"
	pgmodels "quickflow/internal/repository/postgres/postgres-models"
	"quickflow/internal/usecase"
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

	getChatQuery = `
		SELECT id, name, avatar_url, type, created_at, updated_at
		FROM chat
		WHERE id = $1
`

	getPrivateChatQuery = `
		SELECT id, name, avatar_url, type, created_at, updated_at
		FROM chat
		WHERE type = $1 AND id in
			(select cu1.chat_id 
			    from chat_user cu1 join chat_user cu2 on cu1.chat_id = cu2.chat_id
			    where cu1.user_id = $2 and cu2.user_id = $3
		)
`
	getChatParticipantsQuery = `
		SELECT u.id, u.username 
		FROM chat_user cu JOIN "user" u ON cu.user_id = u.id 
		WHERE cu.chat_id = $1
`
)

type ChatRepository struct {
	connPool *pgxpool.Pool
}

func NewPostgresChatRepository() *ChatRepository {
	connPool, err := pgxpool.New(context.Background(), postgres.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &ChatRepository{connPool: connPool}
}

// Close закрывает пул соединений
func (c *ChatRepository) Close() {
	c.connPool.Close()
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

func (c *ChatRepository) GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error) {
	var chatPostgres pgmodels.ChatPostgres
	err := c.connPool.QueryRow(ctx, getChatQuery, chatId).Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL, &chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Chat{}, usecase.ErrNotFound
	} else if err != nil {
		return models.Chat{}, err
	}

	return *chatPostgres.ToChat(), nil
}

func (c *ChatRepository) GetPrivateChat(ctx context.Context, sender, receiver uuid.UUID) (models.Chat, error) {
	var chatPostgres pgmodels.ChatPostgres
	err := c.connPool.QueryRow(ctx, getPrivateChatQuery, models.ChatTypePrivate, sender, receiver).
		Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL,
			&chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Chat{}, usecase.ErrNotFound
	} else if err != nil {
		return models.Chat{}, err
	}

	return *chatPostgres.ToChat(), nil
}

func (c *ChatRepository) Exists(ctx context.Context, chatId uuid.UUID) (bool, error) {
	var exists bool
	err := c.connPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM chat WHERE id = $1)", chatId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (c *ChatRepository) DeleteChat(ctx context.Context, chatId uuid.UUID) error {
	_, err := c.connPool.Exec(ctx, "DELETE FROM chat WHERE id = $1", chatId)
	if err != nil {
		return err
	}
	return nil
}
func (c *ChatRepository) IsParticipant(ctx context.Context, chatId, userId uuid.UUID) (bool, error) {
	var exists bool
	err := c.connPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM chat_user WHERE chat_id = $1 AND user_id = $2)", chatId, userId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *ChatRepository) JoinChat(ctx context.Context, chatId, userId uuid.UUID) error {
	_, err := c.connPool.Exec(ctx, "INSERT INTO chat_user (chat_id, user_id) VALUES ($1, $2)", chatId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChatRepository) LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error {
	_, err := c.connPool.Exec(ctx, "DELETE FROM chat_user WHERE chat_id = $1 AND user_id = $2", chatId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChatRepository) GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]models.User, error) {
	var users []models.User
	rows, err := c.connPool.Query(ctx, getChatParticipantsQuery, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user pgmodels.UserPostgres
		err = rows.Scan(&user.Id, &user.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, user.ConvertToUser())
	}

	if len(users) == 0 {
		return nil, usecase.ErrNotFound
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
