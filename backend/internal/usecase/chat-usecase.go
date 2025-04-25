package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils/validation"
)

var (
	ErrInvalidChatCreationInfo = fmt.Errorf("invalid chat creation info")
	ErrAlreadyInChat           = fmt.Errorf("user already in chat")
	ErrInvalidChatType         = fmt.Errorf("invalid chat type")
)

type ChatRepository interface {
	CreateChat(ctx context.Context, chat models.Chat) error
	GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error)
	GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]models.User, error)
	GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error)
	GetPrivateChat(ctx context.Context, senderId, receiverId uuid.UUID) (models.Chat, error)
	Exists(ctx context.Context, chatId uuid.UUID) (bool, error)
	DeleteChat(ctx context.Context, chatId uuid.UUID) error
	IsParticipant(ctx context.Context, chatId, userId uuid.UUID) (bool, error)
	JoinChat(ctx context.Context, chatId, userId uuid.UUID) error
	LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error
}

type ChatService struct {
	chatRepo    ChatRepository
	fileRepo    FileRepository
	profileRepo ProfileRepository
	messageRepo MessageRepository
}

func NewChatUseCase(charRepo ChatRepository, fileRepo FileRepository, profileRepo ProfileRepository, messageRepo MessageRepository) *ChatService {
	return &ChatService{
		chatRepo:    charRepo,
		fileRepo:    fileRepo,
		profileRepo: profileRepo,
		messageRepo: messageRepo,
	}
}

// CreateChat создает новый чат
func (c *ChatService) CreateChat(ctx context.Context, chatInfo models.ChatCreationInfo) (models.Chat, error) {
	// validation
	if err := validation.ValidateChatCreationInfo(chatInfo); err != nil {
		return models.Chat{}, ErrInvalidChatCreationInfo
	}

	var chat models.Chat
	switch chatInfo.Type {
	case models.ChatTypePrivate:
		chat = models.Chat{
			Type:      chatInfo.Type,
			ID:        uuid.New(),
			AvatarURL: "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	case models.ChatTypeGroup:
		imageURL, err := c.fileRepo.UploadFile(ctx, chatInfo.Avatar)
		if err != nil {
			return models.Chat{}, ErrUploadFile
		}
		chat = models.Chat{
			Type:      chatInfo.Type,
			ID:        uuid.New(),
			Name:      chatInfo.Name,
			AvatarURL: imageURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	if err := c.chatRepo.CreateChat(ctx, chat); err != nil {
		return models.Chat{}, fmt.Errorf("c.chatRepo.CreateChat: %w", err)
	}

	return chat, nil
}

func (c *ChatService) GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error) {
	chats, err := c.chatRepo.GetUserChats(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("c.chatRepo.GetUserChats: %w", err)
	}
	for i := range chats {
		if chats[i].Type == models.ChatTypePrivate {
			chatParticipants, err := c.chatRepo.GetChatParticipants(ctx, chats[i].ID)
			if err != nil {
				return nil, fmt.Errorf("c.profileRepo.GetPublicUserInfo: %w", err)
			}

			chatParticipantsUUIDs := make([]uuid.UUID, len(chatParticipants))
			for j, participant := range chatParticipants {
				chatParticipantsUUIDs[j] = participant.Id
			}
			publicUsersInfo, err := c.profileRepo.GetPublicUsersInfo(ctx, chatParticipantsUUIDs)

			for j := range publicUsersInfo {
				if publicUsersInfo[j].Id != userId {
					chats[i].Name = publicUsersInfo[j].Firstname + " " + publicUsersInfo[j].Lastname
					chats[i].AvatarURL = publicUsersInfo[j].AvatarURL
					break
				}
			}

			lastMessage, err := c.messageRepo.GetLastChatMessage(ctx, chats[i].ID)
			if err != nil {
				return nil, fmt.Errorf("c.messageRepo.GetLastChatMessage: %w", err)
			}

			if lastMessage != nil {
				chats[i].LastMessage = *lastMessage
			}
		}
	}
	return chats, nil
}

func (c *ChatService) GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error) {
	chat, err := c.chatRepo.GetChat(ctx, chatId)
	if err != nil {
		return models.Chat{}, fmt.Errorf("c.chatRepo.GetChat: %w", err)
	}
	return chat, nil
}

func (c *ChatService) DeleteChat(ctx context.Context, chatId uuid.UUID) error {
	exists, err := c.chatRepo.Exists(ctx, chatId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.Exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	err = c.chatRepo.DeleteChat(ctx, chatId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.DeleteChat: %w", err)
	}
	return nil
}

func (c *ChatService) JoinChat(ctx context.Context, chatId, userId uuid.UUID) error {
	exists, err := c.chatRepo.Exists(ctx, chatId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.Exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	isParticipant, err := c.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.IsParticipant: %w", err)
	}
	if isParticipant {
		return ErrAlreadyInChat
	}

	err = c.chatRepo.JoinChat(ctx, chatId, userId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.JoinChat: %w", err)
	}
	return nil
}

func (c *ChatService) LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error {
	exists, err := c.chatRepo.Exists(ctx, chatId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.Exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	isParticipant, err := c.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return ErrNotFound
	}

	err = c.chatRepo.LeaveChat(ctx, chatId, userId)
	if err != nil {
		return fmt.Errorf("c.chatRepo.LeaveChat: %w", err)
	}
	return nil
}

func (c *ChatService) GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]models.User, error) {
	participants, err := c.chatRepo.GetChatParticipants(ctx, chatId)
	if err != nil {
		return nil, fmt.Errorf("c.chatRepo.GetChatParticipants: %w", err)
	}
	return participants, nil
}

func (c *ChatService) GetPrivateChat(ctx context.Context, userId1, userId2 uuid.UUID) (models.Chat, error) {
	chat, err := c.chatRepo.GetPrivateChat(ctx, userId1, userId2)
	if err != nil {
		return models.Chat{}, fmt.Errorf("c.chatRepo.GetPrivateChat: %w", err)
	}
	return chat, nil
}
