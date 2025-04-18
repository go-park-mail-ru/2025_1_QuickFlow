package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "quickflow/internal/models"
    pgmodels "quickflow/internal/repository/postgres/postgres-models"
    "quickflow/internal/usecase"
    "quickflow/pkg/logger"
)

const (
    insertChatQuery = `
        INSERT INTO chat (id, name, avatar_url, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
`
    getUserChatsQuery = `
        SELECT c.id, c.name, c.avatar_url, c.type, c.created_at, c.updated_at, cu.last_read
        FROM chat c
        join chat_user cu on c.id = cu.chat_id
        WHERE cu.user_id = $1
        ORDER BY c.updated_at DESC
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
			    where cu1.user_id = $2 and cu2.user_id = $3)
`

    getLastMessageReadTs = `
		select last_read 
		from chat_user
		where chat_id = $1 and user_id = $2
	`
    getChatParticipantsQuery = `
		SELECT u.id, u.username 
		FROM chat_user cu JOIN "user" u ON cu.user_id = u.id 
		WHERE cu.chat_id = $1
`
)

type ChatRepository struct {
    ConnPool *sql.DB
}

func NewPostgresChatRepository(db *sql.DB) *ChatRepository {
    return &ChatRepository{ConnPool: db}
}

// Close закрывает пул соединений
func (c *ChatRepository) Close() {
    c.ConnPool.Close()
}

func (c *ChatRepository) CreateChat(ctx context.Context, chat models.Chat) error {
    switch chat.Type {
    case models.ChatTypePrivate:
        _, err := c.ConnPool.ExecContext(ctx, insertChatQuery, chat.ID, nil, nil, chat.Type, chat.CreatedAt, chat.UpdatedAt)
        if err != nil {
            logger.Error(ctx, fmt.Sprintf("Unable to save private chat %v to database: %s", chat, err.Error()))
            return err
        }
    case models.ChatTypeGroup:
        _, err := c.ConnPool.ExecContext(ctx, insertChatQuery, chat.ID, chat.Name, chat.AvatarURL, chat.Type, chat.CreatedAt, chat.UpdatedAt)
        if err != nil {
            logger.Error(ctx, fmt.Sprintf("Unable to save group chat %v to database: %s", chat, err.Error()))
            return err
        }
    default:
        logger.Error(ctx, fmt.Sprintf("Invalid chat type %v", chat.Type))
        return usecase.ErrInvalidChatType
    }

    return nil
}

func (c *ChatRepository) GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error) {
    var chats []models.Chat
    rows, err := c.ConnPool.QueryContext(ctx, getUserChatsQuery, userId)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to get user %v chats from database: %s", userId, err.Error()))
        return nil, err
    }
    defer rows.Close()

    var chatPostgres pgmodels.ChatPostgres

    for rows.Next() {
        err = rows.Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL, &chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt, &chatPostgres.LastRead)
        if err != nil {
            logger.Error(ctx, fmt.Sprintf("Unable to scan chat from database for user %v: %s", userId, err.Error()))
            return nil, err
        }
        logger.Info(ctx, fmt.Sprintf("chat name %v last read %v", chatPostgres.Name, chatPostgres.LastRead))
        chats = append(chats, *chatPostgres.ToChat())
    }

    logger.Info(ctx, fmt.Sprintf("Fetched %d chats for user %s", len(chats), userId))
    logger.Info(ctx, fmt.Sprintf("Chats: %#v", chats))

    return chats, nil
}

func (c *ChatRepository) GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error) {
    var chatPostgres pgmodels.ChatPostgres
    err := c.ConnPool.QueryRowContext(ctx, getChatQuery, chatId).Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL, &chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        logger.Error(ctx, fmt.Sprintf("Chat with id %s not found", chatId))
        return models.Chat{}, usecase.ErrNotFound
    } else if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to get chat %v from database: %s", chatId, err.Error()))
        return models.Chat{}, err
    }

    logger.Info(ctx, fmt.Sprintf("Fetched chat %v", chatPostgres))

    return *chatPostgres.ToChat(), nil
}

func (c *ChatRepository) GetPrivateChat(ctx context.Context, requester, companion uuid.UUID) (models.Chat, error) {
    var chatPostgres pgmodels.ChatPostgres
    err := c.ConnPool.QueryRowContext(ctx, getPrivateChatQuery, models.ChatTypePrivate, requester, companion).
        Scan(&chatPostgres.Id, &chatPostgres.Name, &chatPostgres.AvatarURL,
            &chatPostgres.Type, &chatPostgres.CreatedAt, &chatPostgres.UpdatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        logger.Error(ctx, fmt.Sprintf("Private chat between %s and %s not found", requester, companion))
        return models.Chat{}, usecase.ErrNotFound
    } else if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to get private chat between %s and %s from database: %s", requester, companion, err.Error()))
        return models.Chat{}, err
    }

    // get last read ts
    err = c.ConnPool.QueryRowContext(ctx, getLastReadMessageQuery, chatPostgres.Id, companion).Scan(&chatPostgres.LastRead)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to get last read message from database: %s", err.Error()))
        return models.Chat{}, err
    }

    logger.Info(ctx, fmt.Sprintf("Fetched private chat between %s and %s", requester, companion))
    return *chatPostgres.ToChat(), nil
}

func (c *ChatRepository) Exists(ctx context.Context, chatId uuid.UUID) (bool, error) {
    var exists bool
    err := c.ConnPool.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM chat WHERE id = $1)", chatId).Scan(&exists)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to check if chat %v exists: %s", chatId, err.Error()))
        return false, err
    }
    return exists, nil
}
func (c *ChatRepository) DeleteChat(ctx context.Context, chatId uuid.UUID) error {
    _, err := c.ConnPool.ExecContext(ctx, "DELETE FROM chat WHERE id = $1", chatId)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to delete chat %v from database: %s", chatId, err.Error()))
        return err
    }
    return nil
}
func (c *ChatRepository) IsParticipant(ctx context.Context, chatId, userId uuid.UUID) (bool, error) {
    var exists bool
    err := c.ConnPool.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM chat_user WHERE chat_id = $1 AND user_id = $2)", chatId, userId).Scan(&exists)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to check if user %v is participant in chat %v: %s", userId, chatId, err.Error()))
        return false, err
    }
    return exists, nil
}

func (c *ChatRepository) JoinChat(ctx context.Context, chatId, userId uuid.UUID) error {
    _, err := c.ConnPool.ExecContext(ctx, "INSERT INTO chat_user (chat_id, user_id) VALUES ($1, $2)", chatId, userId)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to add user %v to chat %v: %s", userId, chatId, err.Error()))
        return err
    }
    return nil
}

func (c *ChatRepository) LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error {
    _, err := c.ConnPool.ExecContext(ctx, "DELETE FROM chat_user WHERE chat_id = $1 AND user_id = $2", chatId, userId)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to remove user %v from chat %v: %s", userId, chatId, err.Error()))
        return err
    }
    return nil
}

func (c *ChatRepository) GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]models.User, error) {
    var users []models.User
    rows, err := c.ConnPool.QueryContext(ctx, getChatParticipantsQuery, chatId)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Unable to get chat %v participants from database: %s", chatId, err.Error()))
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var user pgmodels.UserPostgres
        err = rows.Scan(&user.Id, &user.Username)
        if err != nil {
            logger.Error(ctx, fmt.Sprintf("Unable to scan user from database for chat %v: %s", chatId, err.Error()))
            return nil, err
        }
        users = append(users, user.ConvertToUser())
    }

    if len(users) == 0 {
        logger.Error(ctx, fmt.Sprintf("No participants found for chat %v", chatId))
        return nil, usecase.ErrNotFound
    }

    if err = rows.Err(); err != nil {
        logger.Error(ctx, fmt.Sprintf("Error while iterating over chat %v participants: %s", chatId, err.Error()))
        return nil, err
    }

    logger.Info(ctx, fmt.Sprintf("Fetched %d participants for chat %s", len(users), chatId))

    return users, nil
}
