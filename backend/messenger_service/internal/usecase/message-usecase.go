package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/shared/models"
)

type MessageRepository interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (models.Message, error)
	GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, numMessages int, timestamp time.Time) ([]models.Message, error)
	GetLastChatMessage(ctx context.Context, chatId uuid.UUID) (*models.Message, error)

	SaveMessage(ctx context.Context, message models.Message) error

	DeleteMessage(ctx context.Context, messageId uuid.UUID) error

	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*time.Time, error)
	UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId uuid.UUID, userId uuid.UUID) error
	GetNumUnreadMessages(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (int, error)
}

type MessageValidator interface {
	ValidateMessage(message models.Message) error
}

type MessageService struct {
	fileRepo    FileService
	messageRepo MessageRepository
	chatRepo    ChatRepository
	validator   MessageValidator
}

func NewMessageService(messageRepo MessageRepository, fileRepo FileService, chatRepo ChatRepository, validator MessageValidator) *MessageService {
	return &MessageService{
		fileRepo:    fileRepo,
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
		validator:   validator,
	}
}

func (m *MessageService) GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, numMessages int, timestamp time.Time) ([]models.Message, error) {
	// validation
	if numMessages <= 0 {
		return nil, messenger_errors.ErrInvalidNumMessages
	}

	// check if user is participant
	isParticipant, err := m.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return nil, fmt.Errorf("m.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return nil, messenger_errors.ErrNotParticipant
	}

	messages, err := m.messageRepo.GetMessagesForChatOlder(ctx, chatId, numMessages, timestamp)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageService) GetLastChatMessage(ctx context.Context, chatId uuid.UUID) (*models.Message, error) {
	// validate
	if chatId == uuid.Nil {
		return nil, fmt.Errorf("chatId is empty")
	}

	message, err := m.messageRepo.GetLastChatMessage(ctx, chatId)
	if err != nil {
		return nil, fmt.Errorf("m.messageRepo.GetLastChatMessage: %w", err)
	}
	return message, nil
}

func (m *MessageService) SaveMessage(ctx context.Context, message models.Message) (*models.Message, error) {
	// validate
	err := m.validator.ValidateMessage(message)
	if err != nil {
		return nil, fmt.Errorf("validation.ValidateMessage: %w", err)
	}

	// check if chat exists and create if it doesn't
	if message.ChatID == uuid.Nil {
		if message.ReceiverID == uuid.Nil {
			return nil, fmt.Errorf("both chatId and receiverId are empty")
		}

		chat, err := m.chatRepo.GetPrivateChat(ctx, message.SenderID, message.ReceiverID)
		if errors.Is(err, messenger_errors.ErrNotFound) {
			newChat := models.Chat{
				Type:      models.ChatTypePrivate,
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = m.chatRepo.CreateChat(ctx, newChat)
			if err != nil {
				return nil, fmt.Errorf("m.chatRepo.CreateChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.SenderID)
			if err != nil {
				return nil, fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.ReceiverID)
			if err != nil {
				m.chatRepo.LeaveChat(ctx, newChat.ID, message.SenderID)
				return nil, fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			message.ChatID = newChat.ID
		} else if err != nil {
			return nil, fmt.Errorf("m.chatRepo.GetChat: %w", err)
		} else {
			message.ChatID = chat.ID
		}
	}

	// Save message to repository
	err = m.messageRepo.SaveMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	newMessage, err := m.messageRepo.GetMessageById(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("m.messageRepo.GetMessageById: %w", err)
	}

	return &newMessage, nil
}

func (m *MessageService) DeleteMessage(ctx context.Context, messageId uuid.UUID) error {
	// validate
	if messageId == uuid.Nil {
		return fmt.Errorf("messageId is empty")
	}

	err := m.messageRepo.DeleteMessage(ctx, messageId)
	if err != nil {
		return fmt.Errorf("m.messageRepo.DeleteMessage: %w", err)
	}

	return nil
}

func (m *MessageService) UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId, userId uuid.UUID) error {
	// check if user is participant
	isParticipant, err := m.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return fmt.Errorf("m.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return messenger_errors.ErrNotParticipant
	}

	err = m.messageRepo.UpdateLastReadTs(ctx, timestamp, chatId, userId)
	if err != nil {
		return fmt.Errorf("m.messageRepo.UpdateLastMessageRead: %w", err)
	}
	return nil
}

func (m *MessageService) GetLastReadTs(ctx context.Context, chatId, userId uuid.UUID) (*time.Time, error) {
	// validate
	if chatId == uuid.Nil {
		return nil, fmt.Errorf("chatId is empty")
	}

	// check if user is participant
	isParticipant, err := m.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return nil, fmt.Errorf("m.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return nil, messenger_errors.ErrNotParticipant
	}

	ts, err := m.messageRepo.GetLastReadTs(ctx, chatId, userId)
	if err != nil {
		return nil, fmt.Errorf("m.messageRepo.GetLastMessageRead: %w", err)
	}
	return ts, nil
}

func (m *MessageService) GetMessageById(ctx context.Context, messageId uuid.UUID) (models.Message, error) {
	// validate
	if messageId == uuid.Nil {
		return models.Message{}, fmt.Errorf("messageId is empty")
	}

	message, err := m.messageRepo.GetMessageById(ctx, messageId)
	if err != nil {
		return models.Message{}, fmt.Errorf("m.messageRepo.GetMessageById: %w", err)
	}
	return message, nil
}

func (m *MessageService) GetNumUnreadMessages(ctx context.Context, chatId, userId uuid.UUID) (int, error) {
	// validate
	if chatId == uuid.Nil || userId == uuid.Nil {
		return 0, fmt.Errorf("chatId or userId is empty")
	}

	// check if user is participant
	isParticipant, err := m.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return 0, fmt.Errorf("m.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return 0, messenger_errors.ErrNotParticipant
	}

	numUnread, err := m.messageRepo.GetNumUnreadMessages(ctx, chatId, userId)
	if err != nil {
		return 0, fmt.Errorf("m.messageRepo.GetNumUnreadMessages: %w", err)
	}
	return numUnread, nil
}
