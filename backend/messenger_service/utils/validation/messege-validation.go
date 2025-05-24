package validation

import (
	"errors"

	"github.com/google/uuid"

	"quickflow/shared/models"
)

type MessageValidator struct{}

func NewMessageValidator() *MessageValidator {
	return &MessageValidator{}
}

func (m *MessageValidator) ValidateMessage(message models.Message) error {
	if len(message.Text) == 0 && len(message.Attachments) == 0 {
		return errors.New("message cannot be empty")
	}
	// TODO make clean, move to config
	if len(message.Attachments) > 10 {
		return errors.New("too many attachments")
	}
	if message.ChatID == uuid.Nil && message.SenderID == uuid.Nil {
		return errors.New("sender ID and chat ID cannot be empty at the same time")
	}
	return nil
}
