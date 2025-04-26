package usecase

import (
	"context"
	"errors"
	"fmt"
	models2 "quickflow/monolith/internal/models"
	"quickflow/monolith/utils/validation"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidNumMessages = fmt.Errorf("numMessages must be greater than 0")
	ErrNotParticipant     = fmt.Errorf("user is not a participant in the chat")
)

type MessageRepository interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (models2.Message, error)
	GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, numMessages int, timestamp time.Time) ([]models2.Message, error)
	GetLastChatMessage(ctx context.Context, chatId uuid.UUID) (*models2.Message, error)

	SaveMessage(ctx context.Context, message models2.Message) error

	DeleteMessage(ctx context.Context, messageId uuid.UUID) error

	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*time.Time, error)
	UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId uuid.UUID, userId uuid.UUID) error
}

type MessageService struct {
	fileRepo    FileRepository
	messageRepo MessageRepository
	chatRepo    ChatRepository
}

func NewMessageService(messageRepo MessageRepository, fileRepo FileRepository, chatRepo ChatRepository) *MessageService {
	return &MessageService{
		fileRepo:    fileRepo,
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

func (m *MessageService) GetMessagesForChat(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, numMessages int, timestamp time.Time) ([]models2.Message, error) {
	// validation
	if numMessages <= 0 {
		return nil, ErrInvalidNumMessages
	}

	// check if user is participant
	isParticipant, err := m.chatRepo.IsParticipant(ctx, chatId, userId)
	if err != nil {
		return nil, fmt.Errorf("m.chatRepo.IsParticipant: %w", err)
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	messages, err := m.messageRepo.GetMessagesForChatOlder(ctx, chatId, numMessages, timestamp)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageService) SaveMessage(ctx context.Context, message models2.Message) (uuid.UUID, error) {
	// validate
	err := validation.ValidateMessage(message)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation.ValidateMessage: %w", err)
	}

	// check if chat exists and create if it doesn't
	if message.ChatID == uuid.Nil {
		if message.ReceiverID == uuid.Nil {
			return uuid.Nil, fmt.Errorf("both chatId and receiverId are empty")
		}

		chat, err := m.chatRepo.GetPrivateChat(ctx, message.SenderID, message.ReceiverID)
		if errors.Is(err, ErrNotFound) {
			newChat := models2.Chat{
				Type:      models2.ChatTypePrivate,
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = m.chatRepo.CreateChat(ctx, newChat)
			if err != nil {
				return uuid.Nil, fmt.Errorf("m.chatRepo.CreateChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.SenderID)
			if err != nil {
				return uuid.Nil, fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.ReceiverID)
			if err != nil {
				m.chatRepo.LeaveChat(ctx, newChat.ID, message.SenderID)
				return uuid.Nil, fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			message.ChatID = newChat.ID
		} else if err != nil {
			return uuid.Nil, fmt.Errorf("m.chatRepo.GetChat: %w", err)
		} else {
			message.ChatID = chat.ID
		}
	}

	// Upload files to storage
	if len(message.Attachments) > 0 {
		filesURLs, err := m.fileRepo.UploadManyFiles(ctx, message.Attachments)
		if err != nil {
			return uuid.Nil, fmt.Errorf("m.fileRepo.UploadManyFiles: %w", err)
		}
		message.AttachmentURLs = filesURLs
	}

	// Save message to repository
	err = m.messageRepo.SaveMessage(ctx, message)
	if err != nil {
		return uuid.Nil, err
	}

	return message.ChatID, nil
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
		return ErrNotParticipant
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
		return nil, ErrNotParticipant
	}

	ts, err := m.messageRepo.GetLastReadTs(ctx, chatId, userId)
	if err != nil {
		return nil, fmt.Errorf("m.messageRepo.GetLastMessageRead: %w", err)
	}
	return ts, nil
}

func (m *MessageService) GetMessageById(ctx context.Context, messageId uuid.UUID) (models2.Message, error) {
	// validate
	if messageId == uuid.Nil {
		return models2.Message{}, fmt.Errorf("messageId is empty")
	}

	message, err := m.messageRepo.GetMessageById(ctx, messageId)
	if err != nil {
		return models2.Message{}, fmt.Errorf("m.messageRepo.GetMessageById: %w", err)
	}
	return message, nil
}
