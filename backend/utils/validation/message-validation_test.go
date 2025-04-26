package validation

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"quickflow/internal/models"
)

func TestValidateMessage(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name     string
		input    models.Message
		expected error
	}{
		{
			name: "valid message with text only",
			input: models.Message{
				Text:     "Hello, world!",
				ChatID:   validUUID,
				SenderID: validUUID,
			},
			expected: nil,
		},
		{
			name: "empty message",
			input: models.Message{
				Text:     "",
				ChatID:   validUUID,
				SenderID: validUUID,
			},
			expected: errors.New("message cannot be empty"),
		},
		{
			name: "message too long",
			input: models.Message{
				Text:     string(make([]rune, 4097)), // 4097 runes
				ChatID:   validUUID,
				SenderID: validUUID,
			},
			expected: errors.New("message too long"),
		},
		{
			name: "too many attachments",
			input: models.Message{
				Text:           "Valid message",
				AttachmentURLs: make([]string, 11),
				ChatID:         validUUID,
				SenderID:       validUUID,
			},
			expected: errors.New("too many attachments"),
		},
		{
			name: "empty chatID and senderID",
			input: models.Message{
				Text:     "Valid message",
				ChatID:   uuid.Nil,
				SenderID: uuid.Nil,
			},
			expected: errors.New("sender ID and chat ID cannot be empty at the same time"),
		},
		{
			name: "valid message with only one ID",
			input: models.Message{
				Text:     "Message with only chat ID",
				ChatID:   validUUID,
				SenderID: uuid.Nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		err := ValidateMessage(tt.input)
		if tt.expected != nil {
			require.EqualError(t, err, tt.expected.Error(), tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}
