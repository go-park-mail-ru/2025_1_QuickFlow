package validation

import (
	"errors"
	models2 "quickflow/monolith/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateChatCreationInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    models2.ChatCreationInfo
		expected error
	}{
		{
			name: "valid private chat",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypePrivate,
			},
			expected: nil,
		},
		{
			name: "private chat with name",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypePrivate,
				Name: "PrivateName",
			},
			expected: errors.New("unexpected name for private chat"),
		},
		{
			name: "private chat with avatar",
			input: models2.ChatCreationInfo{
				Type:   models2.ChatTypePrivate,
				Avatar: &models2.File{Name: "Avatar"},
			},
			expected: errors.New("unexpected avatar for private chat"),
		},
		{
			name: "valid group chat",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypeGroup,
				Name: "GroupName",
			},
			expected: nil,
		},
		{
			name: "group chat with empty name",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypeGroup,
				Name: "",
			},
			expected: errors.New("empty name for group chat"),
		},
		{
			name: "group chat with short name",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypeGroup,
				Name: "Go",
			},
			expected: errors.New("name too short for group chat"),
		},
		{
			name: "group chat with long name",
			input: models2.ChatCreationInfo{
				Type: models2.ChatTypeGroup,
				Name: "ThisGroupChatNameIsWayTooLongToBeValid",
			},
			expected: errors.New("name too long for group chat"),
		},
		{
			name: "invalid chat type",
			input: models2.ChatCreationInfo{
				Type: 999,
			},
			expected: errors.New("invalid chat type"),
		},
	}

	for _, tt := range tests {
		err := ValidateChatCreationInfo(tt.input)
		if tt.expected != nil {
			require.EqualError(t, err, tt.expected.Error(), tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}
