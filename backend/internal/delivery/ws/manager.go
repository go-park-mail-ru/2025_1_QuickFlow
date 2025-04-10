package ws

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "github.com/gorilla/websocket"
    "log"
    "quickflow/internal/delivery/forms"
    http2 "quickflow/internal/delivery/http"
    "quickflow/internal/models"
    "time"
)

// WebSocketConnection представляет собой соединение WebSocket с дополнительными данными о пользователе.
type WebSocketConnection struct {
    UserId   uuid.UUID
    Conn     *websocket.Conn
    LastSeen time.Time
}

// WebSocketManager управляет всеми активными WebSocket соединениями.
type WebSocketManager struct {
    Connections    map[uuid.UUID]*WebSocketConnection
    MessageUseCase http2.MessageUseCase
    ChatUseCase    http2.ChatUseCase
}

// NewWebSocketManager создает новый WebSocketManager.
func NewWebSocketManager(messageUseCase http2.MessageUseCase, chatUseCase http2.ChatUseCase) *WebSocketManager {
    return &WebSocketManager{
        Connections:    make(map[uuid.UUID]*WebSocketConnection),
        MessageUseCase: messageUseCase,
        ChatUseCase:    chatUseCase,
    }
}

// AddConnection добавляет новое соединение в менеджер.
func (wm *WebSocketManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
    wm.Connections[userId] = &WebSocketConnection{
        UserId:   userId,
        Conn:     conn,
        LastSeen: time.Now(),
    }
}

// RemoveConnection удаляет соединение из менеджера.
func (wm *WebSocketManager) RemoveConnection(userId uuid.UUID) {
    if conn, exists := wm.Connections[userId]; exists {
        conn.Conn.Close()
        delete(wm.Connections, userId)
    }
}

// SendMessageToUser отправляет сообщение конкретному пользователю.
func (wm *WebSocketManager) SendMessageToUser(ctx context.Context, userId uuid.UUID, message interface{}) error {
    conn, exists := wm.Connections[userId]
    if !exists {
        return fmt.Errorf("user not connected")
    }

    // Преобразуем сообщение в JSON
    msgJSON, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    err = conn.Conn.WriteMessage(websocket.TextMessage, msgJSON)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    return nil
}

// SendMessageToChat отправляет сообщение всем пользователям чата.
func (wm *WebSocketManager) SendMessageToChat(ctx context.Context, chatId uuid.UUID, message interface{}) error {
    // Получаем список пользователей чата
    chats, err := wm.ChatUseCase.GetChatParticipants(ctx, chatId)
    if err != nil {
        return fmt.Errorf("failed to get users in chat: %w", err)
    }

    for _, chat := range chats {
        err := wm.SendMessageToUser(ctx, chat.Id, message)
        if err != nil {
            log.Println("Failed to send message to user:", chat.Id, err)
        }
    }

    return nil
}

// HandleMessages прослушивает и обрабатывает сообщения от клиента.
func (wm *WebSocketManager) HandleMessages(conn *websocket.Conn, user *models.User) {
    defer conn.Close()

    for {
        // Чтение JSON-сообщения от клиента
        var messageForm forms.MessageForm
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            return
        }

        // Десериализация JSON в структуру MessageForm
        err = json.Unmarshal(msg, &messageForm)
        if err != nil {
            log.Println("Error unmarshaling message:", err)
            return
        }

        // Получаем chatId из сообщения
        chatId := messageForm.ChatId

        // Создаем новое сообщение
        message := messageForm.ToMessageModel()

        // Сохраняем сообщение
        err = wm.MessageUseCase.SaveMessage(context.Background(), message)
        if err != nil {
            log.Println("Failed to save message:", err)
        }

        // Отправляем сообщение всем участникам чата
        err = wm.SendMessageToChat(context.Background(), chatId, message)
        if err != nil {
            log.Println("Failed to send message to chat:", err)
        }
    }
}
