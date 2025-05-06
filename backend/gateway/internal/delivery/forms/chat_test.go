package forms_test

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"quickflow/gateway/internal/delivery/forms"
)

func TestGetChatsForm_GetParams(t *testing.T) {
	tests := []struct {
		name          string
		values        url.Values
		expectedForm  forms.GetChatsForm
		expectedError error
	}{
		{
			name: "success with valid parameters",
			values: url.Values{
				"chats_count": []string{"5"},
				"ts":          []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm: forms.GetChatsForm{
				ChatsCount: 5,
				Ts:         time.Date(2025, 4, 16, 0, 0, 0, 0, time.UTC),
			},
			expectedError: nil,
		},
		{
			name: "missing chats_count parameter",
			values: url.Values{
				"ts": []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms.GetChatsForm{},
			expectedError: errors.New("chats_count parameter missing"),
		},
		{
			name: "invalid chats_count format",
			values: url.Values{
				"chats_count": []string{"invalid"},
				"ts":          []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms.GetChatsForm{},
			expectedError: errors.New("failed to parse chats_count"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form forms.GetChatsForm
			err := form.GetParams(tt.values)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				// We compare the dates carefully because time.Now() is dynamic.
				if !tt.expectedForm.Ts.Equal(form.Ts) {
					t.Errorf("expected Ts %v, got %v", tt.expectedForm.Ts, form.Ts)
				}
				assert.Equal(t, tt.expectedForm.ChatsCount, form.ChatsCount)
			}
		})
	}
}
