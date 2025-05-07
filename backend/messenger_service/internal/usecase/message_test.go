package usecase

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"quickflow/shared/models"
	"testing"
	"time"
)

func TestGetMessagesForChatOlder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repository.NewMockMessageRepository(ctrl)
	validator := validator.NewMockMessageValidator(ctrl)
	service := service.NewMessageService(repo, validator)

	ctx := context.Background()
	chatID := int64(1)
	limit := 10
	before := time.Now()

	messages := []models.Message{
		{ID: 1, ChatID: chatID, SenderID: 2, Text: "Hello", CreatedAt: before.Add(-time.Minute)},
		{ID: 2, ChatID: chatID, SenderID: 3, Text: "Hi", CreatedAt: before.Add(-time.Minute * 2)},
	}

	repo.EXPECT().GetMessagesForChatOlder(ctx, chatID, before, limit).Return(messages, nil)

	result, err := service.GetMessagesForChatOlder(ctx, chatID, before, limit)

	assert.NoError(t, err)
	assert.Equal(t, messages, result)
}
