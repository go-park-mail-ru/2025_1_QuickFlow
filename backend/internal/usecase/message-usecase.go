package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils/validation"
)

var (
	ErrInvalidNumMessages = fmt.Errorf("numMessages must be greater than 0")
	ErrNotParticipant     = fmt.Errorf("user is not a participant in the chat")
)

type MessageRepository interface {
	GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Message, error)
	SaveMessage(ctx context.Context, message models.Message) error
	DeleteMessage(ctx context.Context, messageId uuid.UUID) error
	MarkRead(ctx context.Context, messageId uuid.UUID) error
}

type MessageUseCase struct {
	fileRepo    FileRepository
	messageRepo MessageRepository
	chatRepo    ChatRepository
}

func NewMessageUseCase(messageRepo MessageRepository, fileRepo FileRepository, chatRepo ChatRepository) *MessageUseCase {
	return &MessageUseCase{
		fileRepo:    fileRepo,
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

func (m *MessageUseCase) GetMessagesForChat(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, numMessages int, timestamp time.Time) ([]models.Message, error) {
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

func (m *MessageUseCase) SaveMessage(ctx context.Context, message models.Message) error {
	// validate
	err := validation.ValidateMessage(message)
	if err != nil {
		return fmt.Errorf("validation.ValidateMessage: %w", err)
	}

	// check if chat exists and create if it doesn't
	if message.ChatID == uuid.Nil {
		if message.ReceiverID == uuid.Nil {
			return fmt.Errorf("both chatId and receiverId are empty")
		}

		chat, err := m.chatRepo.GetPrivateChat(ctx, message.SenderID, message.ReceiverID)
		if errors.Is(err, ErrNotFound) {
			newChat := models.Chat{
				Type:      models.ChatTypePrivate,
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = m.chatRepo.CreateChat(ctx, newChat)
			if err != nil {
				return fmt.Errorf("m.chatRepo.CreateChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.SenderID)
			if err != nil {
				return fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			err = m.chatRepo.JoinChat(ctx, newChat.ID, message.ReceiverID)
			if err != nil {
				m.chatRepo.LeaveChat(ctx, newChat.ID, message.SenderID)
				return fmt.Errorf("m.chatRepo.JoinChat: %w", err)
			}
			message.ChatID = newChat.ID
		} else if err != nil {
			return fmt.Errorf("m.chatRepo.GetChat: %w", err)
		} else {
			message.ChatID = chat.ID
		}
	}

	// Upload files to storage
	if len(message.Attachments) > 0 {
		filesURLs, err := m.fileRepo.UploadManyFiles(ctx, message.Attachments)
		if err != nil {
			return fmt.Errorf("m.fileRepo.UploadManyFiles: %w", err)
		}
		message.AttachmentURLs = filesURLs
	}

	// Save message to repository
	err = m.messageRepo.SaveMessage(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageUseCase) DeleteMessage(ctx context.Context, messageId uuid.UUID) error {
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

func (m *MessageUseCase) MarkRead(ctx context.Context, messageId uuid.UUID) error {
	// validate
	if messageId == uuid.Nil {
		return fmt.Errorf("messageId is empty")
	}

	err := m.messageRepo.MarkRead(ctx, messageId)
	if err != nil {
		return fmt.Errorf("m.messageRepo.MarkRead: %w", err)
	}

	return nil
}
