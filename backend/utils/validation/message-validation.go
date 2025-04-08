package validation

import (
	"errors"
	"github.com/google/uuid"
	"quickflow/internal/models"
)

func ValidateMessage(message models.Message) error {
	if len(message.Text) == 0 {
		return errors.New("message cannot be empty")
	}
	if len(message.Text) > 4096 {
		return errors.New("message too long")
	}
	if len(message.AttachmentURLs) > 10 {
		return errors.New("too many attachments")
	}
	if message.ChatID == uuid.Nil {
		return errors.New("chat ID cannot be empty")
	}
	if message.SenderID == uuid.Nil {
		return errors.New("sender ID cannot be empty")
	}
	return nil
}
