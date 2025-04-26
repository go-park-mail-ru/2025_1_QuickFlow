package usecase

import (
	"context"
	models2 "quickflow/monolith/internal/models"
	mocks2 "quickflow/monolith/internal/usecase/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateChat_InvalidChatCreationInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mocks2.NewMockChatRepository(ctrl)
	mockFileRepo := mocks2.NewMockFileRepository(ctrl)
	mockProfileRepo := mocks2.NewMockProfileRepository(ctrl)
	mockMessageRepo := mocks2.NewMockMessageRepository(ctrl)

	usecase := NewChatUseCase(mockChatRepo, mockFileRepo, mockProfileRepo, mockMessageRepo)

	// Создаем неправильные данные для чата (например, имя пустое)
	chatInfo := models2.ChatCreationInfo{
		Type:   models2.ChatTypeGroup,
		Name:   "",
		Avatar: nil,
	}

	// Мокируем валидацию, которая должна вернуть ошибку
	mockFileRepo.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Times(0) // Не будет вызвано, так как валидатор не прошел.

	// Проверяем ошибку
	chat, err := usecase.CreateChat(context.Background(), chatInfo)
	assert.EqualError(t, err, ErrInvalidChatCreationInfo.Error())
	assert.Equal(t, models2.Chat{}, chat)
}

func TestCreateChat_GroupChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mocks2.NewMockChatRepository(ctrl)
	mockFileRepo := mocks2.NewMockFileRepository(ctrl)
	mockProfileRepo := mocks2.NewMockProfileRepository(ctrl)
	mockMessageRepo := mocks2.NewMockMessageRepository(ctrl)

	usecase := NewChatUseCase(mockChatRepo, mockFileRepo, mockProfileRepo, mockMessageRepo)

	// Создаем данные для нового группового чата
	chatInfo := models2.ChatCreationInfo{
		Type:   models2.ChatTypeGroup,
		Name:   "Group Chat",
		Avatar: nil,
	}

	// Мокируем загрузку аватара
	mockFileRepo.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return("http://example.com/avatar.jpg", nil)

	// Мокируем создание чата в репозитории
	mockChatRepo.EXPECT().CreateChat(gomock.Any(), gomock.Any()).Return(nil)

	// Вызываем метод и проверяем результат
	chat, err := usecase.CreateChat(context.Background(), chatInfo)
	assert.NoError(t, err)
	assert.NotNil(t, chat.ID)
	assert.Equal(t, chat.Name, "Group Chat")
	assert.Equal(t, chat.AvatarURL, "http://example.com/avatar.jpg")
}

func TestGetUserChats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mocks2.NewMockChatRepository(ctrl)
	mockFileRepo := mocks2.NewMockFileRepository(ctrl)
	mockProfileRepo := mocks2.NewMockProfileRepository(ctrl)
	mockMessageRepo := mocks2.NewMockMessageRepository(ctrl)

	usecase := NewChatUseCase(mockChatRepo, mockFileRepo, mockProfileRepo, mockMessageRepo)

	userId := uuid.New()

	// Мокируем репозиторий, чтобы вернуть список чатов
	mockChatRepo.EXPECT().GetUserChats(gomock.Any(), userId).Return([]models2.Chat{
		{ID: uuid.New(), Type: models2.ChatTypeGroup, Name: "Group Chat", CreatedAt: time.Now(), LastMessage: models2.Message{Text: "hi"}},
		{ID: uuid.New(), Type: models2.ChatTypePrivate, CreatedAt: time.Now()},
	}, nil)

	// Мокируем получение участников чата для приватного чата
	mockChatRepo.EXPECT().GetChatParticipants(gomock.Any(), gomock.Any()).Return([]models2.User{
		{Id: userId},
	}, nil)

	// Мокируем получение информации о публичных пользователях
	mockProfileRepo.EXPECT().GetPublicUsersInfo(gomock.Any(), gomock.Any()).Return([]models2.PublicUserInfo{
		{Id: userId, Firstname: "John", Lastname: "Doe", AvatarURL: "http://example.com/avatar.jpg"},
	}, nil)

	// Мокируем последние сообщения чата
	mockMessageRepo.EXPECT().GetLastChatMessage(gomock.Any(), gomock.Any()).Return(&models2.Message{Text: "Hello!"}, nil)

	// Проверяем результат
	chats, err := usecase.GetUserChats(context.Background(), userId)
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, chats[0].Name, "Group Chat")
	assert.NotEmpty(t, chats[0].LastMessage.Text)
}

//func TestJoinChat_UserAlreadyInChat(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockChatRepo := mocks.NewMockChatRepository(ctrl)
//	mockFileRepo := mocks.NewMockFileRepository(ctrl)
//	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
//	mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
//
//	usecase := NewChatUseCase(mockChatRepo, mockFileRepo, mockProfileRepo, mockMessageRepo)
//
//	chatId := uuid.New()
//	userId := uuid.New()
//
//	// Мокируем, что пользователь уже в чате
//	mockChatRepo.EXPECT().IsParticipant(gomock.Any(), chatId, userId).Return(true, nil)
//
//	// Проверяем ошибку
//	err := usecase.JoinChat(context.Background(), chatId, userId)
//	assert.EqualError(t, err, ErrAlreadyInChat.Error())
//}
