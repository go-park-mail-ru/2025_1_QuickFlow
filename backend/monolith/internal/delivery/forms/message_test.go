package forms_test

import (
	"errors"
	"net/url"
	forms2 "quickflow/monolith/internal/delivery/forms"
	"quickflow/monolith/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/internal/delivery/forms"
)

func TestGetMessagesForm_GetParams(t *testing.T) {
	tests := []struct {
		name          string
		values        url.Values
		expectedForm  forms2.GetMessagesForm
		expectedError error
	}{
		{
			name: "success with valid parameters",
			values: url.Values{
				"messages_count": []string{"10"},
				"ts":             []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm: forms2.GetMessagesForm{
				MessagesCount: 10,
				Ts:            time.Date(2025, 4, 16, 0, 0, 0, 0, time.UTC),
			},
			expectedError: nil,
		},
		{
			name: "missing messages_count parameter",
			values: url.Values{
				"ts": []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms2.GetMessagesForm{},
			expectedError: errors.New("messages_count parameter missing"),
		},
		{
			name: "invalid messages_count format",
			values: url.Values{
				"messages_count": []string{"invalid"},
				"ts":             []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms2.GetMessagesForm{},
			expectedError: errors.New("failed to parse messages_count"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form forms2.GetMessagesForm
			err := form.GetParams(tt.values)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				// We compare the dates carefully because time.Now() is dynamic.
				if !tt.expectedForm.Ts.Equal(form.Ts) {
					t.Errorf("expected Ts %v, got %v", tt.expectedForm.Ts, form.Ts)
				}
				assert.Equal(t, tt.expectedForm.MessagesCount, form.MessagesCount)
			}
		})
	}
}

func TestToMessageOut(t *testing.T) {
	userID := uuid.New()
	messageID := uuid.New()
	now := time.Now()

	// Create a mock message
	message := models.Message{
		ID:             messageID,
		Text:           "Hello!",
		CreatedAt:      now,
		UpdatedAt:      now,
		IsRead:         true,
		AttachmentURLs: []string{"http://example.com/image.jpg"},
		SenderID:       userID,
		ChatID:         uuid.New(),
	}

	// Create a mock user info
	userInfo := models.PublicUserInfo{
		Id:        userID,
		Username:  "user1",
		Firstname: "John",
		Lastname:  "Doe",
		AvatarURL: "http://example.com/avatar.jpg",
	}

	messageOut := forms2.ToMessageOut(message, userInfo)

	assert.Equal(t, messageOut.ID, messageID)
	assert.Equal(t, messageOut.Text, "Hello!")
	assert.Equal(t, messageOut.IsRead, true)
	assert.Equal(t, messageOut.AttachmentURLs, []string{"http://example.com/image.jpg"})
	assert.Equal(t, messageOut.Sender.ID, userID.String())
	assert.Equal(t, messageOut.Sender.Username, "user1")
}

func TestToMessagesOut(t *testing.T) {
	userID := uuid.New()
	messageID := uuid.New()
	chatID := uuid.New()
	now := time.Now()

	// Create a mock message
	messages := []models.Message{
		{
			ID:             messageID,
			Text:           "Hello!",
			CreatedAt:      now,
			UpdatedAt:      now,
			IsRead:         true,
			AttachmentURLs: []string{"http://example.com/image.jpg"},
			SenderID:       userID,
			ChatID:         chatID,
		},
	}

	// Create a mock user info
	usersInfo := map[uuid.UUID]models.PublicUserInfo{
		userID: {
			Id:        userID,
			Username:  "user1",
			Firstname: "John",
			Lastname:  "Doe",
			AvatarURL: "http://example.com/avatar.jpg",
		},
	}

	messagesOut := forms2.ToMessagesOut(messages, usersInfo)

	assert.Len(t, messagesOut, 1)
	assert.Equal(t, messagesOut[0].ID, messageID)
	assert.Equal(t, messagesOut[0].Text, "Hello!")
	assert.Equal(t, messagesOut[0].Sender.Username, "user1")
	assert.Equal(t, messagesOut[0].AttachmentURLs, []string{"http://example.com/image.jpg"})
}

func TestMessageForm_ToMessageModel(t *testing.T) {
	senderID := uuid.New()
	receiverID := uuid.New()
	chatID := uuid.New()

	messageForm := forms2.MessageForm{
		Text:            "Sample message",
		AttachmentsUrls: []string{"http://example.com/image.jpg"},
		ReceiverId:      receiverID,
		SenderId:        senderID,
		ChatId:          chatID,
	}

	messageModel := messageForm.ToMessageModel()

	assert.Equal(t, messageModel.Text, "Sample message")
	assert.Equal(t, messageModel.ReceiverID, receiverID)
	assert.Equal(t, messageModel.SenderID, senderID)
	assert.Equal(t, messageModel.ChatID, chatID)
	assert.Len(t, messageModel.AttachmentURLs, 1)
	assert.Equal(t, messageModel.AttachmentURLs[0], "http://example.com/image.jpg")
}

func TestMessageRequest(t *testing.T) {
	senderID := uuid.New()
	receiverID := uuid.New()
	chatID := uuid.New()

	messageForm := forms2.MessageForm{
		Text:            "Another message",
		AttachmentsUrls: []string{"http://example.com/image2.jpg"},
		ReceiverId:      receiverID,
		SenderId:        senderID,
		ChatId:          chatID,
	}

	messageRequest := forms.MessageRequest{
		Type:    "text",
		Payload: messageForm,
	}

	assert.Equal(t, messageRequest.Type, "text")
	assert.Equal(t, messageRequest.Payload.Text, "Another message")
	assert.Equal(t, messageRequest.Payload.ReceiverId, receiverID)
	assert.Equal(t, messageRequest.Payload.SenderId, senderID)
}
