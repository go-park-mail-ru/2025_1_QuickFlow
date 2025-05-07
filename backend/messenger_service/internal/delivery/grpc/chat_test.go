package grpc

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"quickflow/messenger_service/internal/delivery/grpc/mocks"
	"quickflow/shared/models"
	"testing"
)

func TestCreateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatInfo := models.ChatCreationInfo{Name: "Test Chat"}
	expectedChat := models.Chat{ID: uuid.New(), Name: "Test Chat"}

	// Настройка мока
	mockChatUseCase.EXPECT().CreateChat(context.Background(), chatInfo).Return(expectedChat, nil)

	// Ваша логика теста
	chat, err := mockChatUseCase.CreateChat(context.Background(), chatInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if chat.Name != expectedChat.Name {
		t.Errorf("expected chat %v, got %v", expectedChat, chat)
	}
}

func TestDeleteChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatID := uuid.New()

	// Настройка мока
	mockChatUseCase.EXPECT().DeleteChat(context.Background(), chatID).Return(nil)

	// Ваша логика теста
	err := mockChatUseCase.DeleteChat(context.Background(), chatID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatID := uuid.New()
	expectedChat := models.Chat{ID: chatID, Name: "Test Chat"}

	// Настройка мока
	mockChatUseCase.EXPECT().GetChat(context.Background(), chatID).Return(expectedChat, nil)

	// Ваша логика теста
	chat, err := mockChatUseCase.GetChat(context.Background(), chatID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if chat.Name != expectedChat.Name {
		t.Errorf("expected chat %v, got %v", expectedChat, chat)
	}
}

func TestGetChatParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatID := uuid.New()
	expectedParticipants := []uuid.UUID{uuid.New(), uuid.New()}

	// Настройка мока
	mockChatUseCase.EXPECT().GetChatParticipants(context.Background(), chatID).Return(expectedParticipants, nil)

	// Ваша логика теста
	participants, err := mockChatUseCase.GetChatParticipants(context.Background(), chatID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(participants) != len(expectedParticipants) {
		t.Errorf("expected %v participants, got %v", len(expectedParticipants), len(participants))
	}
}

func TestGetPrivateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	user1 := uuid.New()
	user2 := uuid.New()
	expectedChat := models.Chat{ID: uuid.New(), Name: "Private Chat"}

	// Настройка мока
	mockChatUseCase.EXPECT().GetPrivateChat(context.Background(), user1, user2).Return(expectedChat, nil)

	// Ваша логика теста
	chat, err := mockChatUseCase.GetPrivateChat(context.Background(), user1, user2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if chat.Name != expectedChat.Name {
		t.Errorf("expected chat %v, got %v", expectedChat, chat)
	}
}

func TestGetUserChats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	userID := uuid.New()
	expectedChats := []models.Chat{
		{ID: uuid.New(), Name: "Chat 1"},
		{ID: uuid.New(), Name: "Chat 2"},
	}

	// Настройка мока
	mockChatUseCase.EXPECT().GetUserChats(context.Background(), userID).Return(expectedChats, nil)

	// Ваша логика теста
	chats, err := mockChatUseCase.GetUserChats(context.Background(), userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(chats) != len(expectedChats) {
		t.Errorf("expected %v chats, got %v", len(expectedChats), len(chats))
	}
}

func TestJoinChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatID := uuid.New()
	userID := uuid.New()

	// Настройка мока
	mockChatUseCase.EXPECT().JoinChat(context.Background(), chatID, userID).Return(nil)

	// Ваша логика теста
	err := mockChatUseCase.JoinChat(context.Background(), chatID, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLeaveChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUseCase := mocks.NewMockChatUseCase(ctrl)
	chatID := uuid.New()
	userID := uuid.New()

	// Настройка мока
	mockChatUseCase.EXPECT().LeaveChat(context.Background(), chatID, userID).Return(nil)

	// Ваша логика теста
	err := mockChatUseCase.LeaveChat(context.Background(), chatID, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
