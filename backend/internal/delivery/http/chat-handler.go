package http

import (
	"context"
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type ChatUseCase interface {
	CreateChat(ctx context.Context, chatInfo models.ChatCreationInfo) (models.Chat, error)
	GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error)
	DeleteChat(ctx context.Context, chatId uuid.UUID) error
	GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error)
	JoinChat(ctx context.Context, chatId, userId uuid.UUID) error
	LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error
}

type ChatHandler struct {
	chatUseCase ChatUseCase
}

func NewChatHandler(chatUseCase ChatUseCase) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
	}
}
