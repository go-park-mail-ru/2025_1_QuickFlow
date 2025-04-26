package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	time2 "quickflow/monolith/config/time"
	"quickflow/monolith/internal/delivery/forms"
	http2 "quickflow/monolith/internal/delivery/http"
	forms2 "quickflow/monolith/internal/delivery/ws/forms"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/pkg/logger"
	"quickflow/monolith/utils/validation"
)

type WSConnectionManager struct {
	Connections map[uuid.UUID]*websocket.Conn
	mu          sync.RWMutex
}

func NewWSConnectionManager() *WSConnectionManager {
	return &WSConnectionManager{
		Connections: make(map[uuid.UUID]*websocket.Conn),
	}
}

// AddConnection adds a new user connection to the manager
func (wm *WSConnectionManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
	wm.mu.Lock()
	wm.Connections[userId] = conn
	wm.mu.Unlock()
}

// RemoveAndCloseConnection removes a user connection from the manager and closes it
func (wm *WSConnectionManager) RemoveAndCloseConnection(userId uuid.UUID) {
	wm.mu.Lock()
	if _, exists := wm.Connections[userId]; exists {
		delete(wm.Connections, userId)
	}
	wm.mu.Unlock()
}

func (wm *WSConnectionManager) IsConnected(userId uuid.UUID) (*websocket.Conn, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	conn, exists := wm.Connections[userId]
	return conn, exists
}

// ---------------------------------------------------------

type InternalWSMessageHandler struct {
	WSConnectionManager *WSConnectionManager
	MessageUseCase      http2.MessageUseCase
	profileUseCase      http2.ProfileUseCase
	ChatUseCase         http2.ChatUseCase
}

func NewInternalWSMessageHandler(wsConnManager *WSConnectionManager, messageUseCase http2.MessageUseCase, profileUseCase http2.ProfileUseCase, chatUseCase http2.ChatUseCase) *InternalWSMessageHandler {
	return &InternalWSMessageHandler{
		WSConnectionManager: wsConnManager,
		MessageUseCase:      messageUseCase,
		profileUseCase:      profileUseCase,
		ChatUseCase:         chatUseCase,
	}
}

func (m *InternalWSMessageHandler) Handle(ctx context.Context, user models.User, payload json.RawMessage) error {
	var messageForm forms.MessageForm
	if err := json.Unmarshal(payload, &messageForm); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	messageForm.SenderId = user.Id
	if messageForm.ChatId == uuid.Nil && messageForm.ReceiverId == uuid.Nil {
		logger.Error(ctx, "ChatId and ReceiverId cannot be both nil")
		return fmt.Errorf("chatId and receiverId cannot be both nil")
	}

	message := messageForm.ToMessageModel()
	if err := validation.ValidateMessage(message); err != nil {
		logger.Error(ctx, "Invalid message:", err)
		return fmt.Errorf("invalid message: %w", err)
	}

	var err error
	message.ChatID, err = m.MessageUseCase.SaveMessage(ctx, message)
	if err != nil {
		log.Println("Failed to save message:", err)
		return fmt.Errorf("failed to save message: %w", err)
	}

	// retrieving info to send message to all chat users
	publicSenderInfo, err := m.profileUseCase.GetPublicUserInfo(ctx, user.Id)
	if err != nil {
		log.Println("Failed to get public sender info:", err)
		return fmt.Errorf("failed to get public sender info: %w", err)
	}
	chatParticipants, err := m.ChatUseCase.GetChatParticipants(ctx, message.ChatID)
	if err != nil {
		log.Println("Failed to get chat participants:", err)
		return fmt.Errorf("failed to get chat participants: %w", err)
	}
	err = m.sendMessageToChat(ctx, message, publicSenderInfo, chatParticipants)
	if err != nil {
		log.Println("Failed to send message to chat:", err)
		return fmt.Errorf("failed to send message to chat: %w", err)
	}

	return nil
}

// SendMessageToUser sends a message to a specific user
func (m *InternalWSMessageHandler) SendMessageToUser(_ context.Context, userId uuid.UUID, message forms.MessageOut) error {
	conn, exists := m.WSConnectionManager.IsConnected(userId)

	if !exists {
		return fmt.Errorf("user not connected")
	}

	out := struct {
		Type    string           `json:"type"`
		Message forms.MessageOut `json:"payload"`
	}{"message", message}

	msgJSON, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// SendMessageToChat sends a message to all participants in a chat
func (m *InternalWSMessageHandler) sendMessageToChat(ctx context.Context, message models.Message, publicSenderInfo models.PublicUserInfo, chatParticipants []models.User) error {
	for _, user := range chatParticipants {
		err := m.SendMessageToUser(ctx, user.Id, forms.ToMessageOut(message, publicSenderInfo))
		if err != nil {
			log.Println("Failed to send message to user:", user.Id, err)
		}
	}

	return nil
}

func (m *InternalWSMessageHandler) MarkMessageRead(ctx context.Context, user models.User, jsonPayload json.RawMessage) error {
	var payload forms2.MarkReadPayload

	if err := json.Unmarshal(jsonPayload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.MessageId == uuid.Nil || payload.ChatId == uuid.Nil {
		return fmt.Errorf("messageId or chatId is empty")
	}

	msg, err := m.MessageUseCase.GetMessageById(ctx, payload.MessageId)
	if err != nil {
		return fmt.Errorf("failed to get message by id: %w", err)
	}

	err = m.MessageUseCase.UpdateLastReadTs(ctx, msg.CreatedAt, payload.ChatId, user.Id)
	if err != nil {
		return fmt.Errorf("failed to update last message read: %w", err)
	}

	// send message to message author
	messageReadForm := forms2.NotifyMessageRead{
		MessageId: payload.MessageId,
		Timestamp: msg.CreatedAt.Format(time2.TimeStampLayout),
		ChatId:    payload.ChatId,
		SenderId:  user.Id,
	}

	err = m.notifyMessageRead(ctx, messageReadForm, msg.SenderID)
	if err != nil {
		return fmt.Errorf("failed to notify message read: %w", err)
	}
	return nil
}

func (m *InternalWSMessageHandler) notifyMessageRead(_ context.Context, read forms2.NotifyMessageRead, receiver uuid.UUID) error {
	conn, exists := m.WSConnectionManager.IsConnected(receiver)
	if !exists {
		return fmt.Errorf("user not connected")
	}

	out := struct {
		Type string                   `json:"type"`
		Data forms2.NotifyMessageRead `json:"payload"`
	}{"message_read", read}

	msgJSON, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

type PingHandler interface {
	Handle(ctx context.Context, conn *websocket.Conn)
}

// PingHandlerWS - Обработчик Ping сообщений
type PingHandlerWS struct{}

func NewPingHandlerWS() *PingHandlerWS {
	return &PingHandlerWS{}
}

func (wm *PingHandlerWS) Handle(ctx context.Context, conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		logger.Info(ctx, "Received pong:", appData)
		return nil
	})

	go func() {
		for {
			time.Sleep(30 * time.Second) // отправка ping каждые 30 секунд
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Info(ctx, "Failed to send ping:", err)
				return
			}
		}
	}()
	return
}
