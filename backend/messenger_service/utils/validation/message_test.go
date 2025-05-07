package validation_test

import (
	"quickflow/messenger_service/utils/validation"
	"testing"

	"github.com/google/uuid"
	"quickflow/shared/models"
)

func TestMessageValidator_ValidateMessage(t *testing.T) {
	validator := validation.NewMessageValidator()

	tests := []struct {
		name        string
		message     models.Message
		expectError bool
	}{
		{
			name: "valid message",
			message: models.Message{
				Text:     "Hello, world!",
				ChatID:   uuid.New(),
				SenderID: uuid.New(),
			},
			expectError: false,
		},
		{
			name: "empty message text",
			message: models.Message{
				ChatID:   uuid.New(),
				SenderID: uuid.New(),
			},
			expectError: true,
		},
		{
			name: "too many attachments",
			message: models.Message{
				Text:           "Message with too many attachments",
				ChatID:         uuid.New(),
				SenderID:       uuid.New(),
				AttachmentURLs: make([]string, 11),
			},
			expectError: true,
		},
		{
			name: "missing chat ID and sender ID",
			message: models.Message{
				Text: "Message without IDs",
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateMessage(test.message)
			if (err != nil) != test.expectError {
				t.Errorf("expected error: %v, got: %v", test.expectError, err)
			}
		})
	}
}
