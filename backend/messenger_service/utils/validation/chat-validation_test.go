package validation

import (
	"errors"
	"quickflow/gateway/utils/validation"
	"testing"

	"quickflow/shared/models"

	"github.com/stretchr/testify/require"
)

func TestValidateChatCreationInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    models.ChatCreationInfo
		expected error
	}{
		{
			name: "valid private chat",
			input: models.ChatCreationInfo{
				Type: models.ChatTypePrivate,
			},
			expected: nil,
		},
		{
			name: "private chat with name",
			input: models.ChatCreationInfo{
				Type: models.ChatTypePrivate,
				Name: "PrivateName",
			},
			expected: errors.New("unexpected name for private chat"),
		},
		{
			name: "private chat with avatar",
			input: models.ChatCreationInfo{
				Type:   models.ChatTypePrivate,
				Avatar: &models.File{Name: "Avatar"},
			},
			expected: errors.New("unexpected avatar for private chat"),
		},
		{
			name: "valid group chat",
			input: models.ChatCreationInfo{
				Type: models.ChatTypeGroup,
				Name: "GroupName",
			},
			expected: nil,
		},
		{
			name: "group chat with empty name",
			input: models.ChatCreationInfo{
				Type: models.ChatTypeGroup,
				Name: "",
			},
			expected: errors.New("empty name for group chat"),
		},
		{
			name: "group chat with short name",
			input: models.ChatCreationInfo{
				Type: models.ChatTypeGroup,
				Name: "Go",
			},
			expected: errors.New("name too short for group chat"),
		},
		{
			name: "group chat with long name",
			input: models.ChatCreationInfo{
				Type: models.ChatTypeGroup,
				Name: "ThisGroupChatNameIsWayTooLongToBeValid",
			},
			expected: errors.New("name too long for group chat"),
		},
		{
			name: "invalid chat type",
			input: models.ChatCreationInfo{
				Type: 999,
			},
			expected: errors.New("invalid chat type"),
		},
	}

	for _, tt := range tests {
		err := validation.ValidateChatCreationInfo(tt.input)
		if tt.expected != nil {
			require.EqualError(t, err, tt.expected.Error(), tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}
