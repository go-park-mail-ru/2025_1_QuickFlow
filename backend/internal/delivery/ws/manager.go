package ws

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/gorilla/websocket"

    "quickflow/internal/delivery/forms"
    http2 "quickflow/internal/delivery/http"
    "quickflow/internal/models"
    "quickflow/pkg/logger"
)

type WebSocketConnection struct {
    UserId   uuid.UUID
    Conn     *websocket.Conn
    LastSeen time.Time
}

type WebSocketManager struct {
    Connections    map[uuid.UUID]*WebSocketConnection
    MessageUseCase http2.MessageUseCase
    ChatUseCase    http2.ChatUseCase
}

func NewWebSocketManager(messageUseCase http2.MessageUseCase, chatUseCase http2.ChatUseCase) *WebSocketManager {
    return &WebSocketManager{
        Connections:    make(map[uuid.UUID]*WebSocketConnection),
        MessageUseCase: messageUseCase,
        ChatUseCase:    chatUseCase,
    }
}

// AddConnection adds a new user connection to the manager
func (wm *WebSocketManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
    wm.Connections[userId] = &WebSocketConnection{
        UserId:   userId,
        Conn:     conn,
        LastSeen: time.Now(),
    }
}

// RemoveConnection removes a user connection from the manager
func (wm *WebSocketManager) RemoveConnection(userId uuid.UUID) {
    if conn, exists := wm.Connections[userId]; exists {
        conn.Conn.Close()
        delete(wm.Connections, userId)
    }
}

// SendMessageToUser sends a message to a specific user
func (wm *WebSocketManager) SendMessageToUser(ctx context.Context, userId uuid.UUID, message interface{}) error {
    conn, exists := wm.Connections[userId]
    if !exists {
        return fmt.Errorf("user not connected")
    }

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

// SendMessageToChat sends a message to all participants in a chat
func (wm *WebSocketManager) SendMessageToChat(ctx context.Context, chatId uuid.UUID, message models.Message) error {
    chats, err := wm.ChatUseCase.GetChatParticipants(ctx, chatId)
    if err != nil {
        return fmt.Errorf("failed to get users in chat: %w", err)
    }

    for _, chat := range chats {
        err := wm.SendMessageToUser(ctx, chat.Id, forms.ToMessageOut(message))
        if err != nil {
            log.Println("Failed to send message to user:", chat.Id, err)
        }
    }

    return nil
}

// HandleMessages godoc
// @Summary Handle incoming messages
// @Description Handle incoming messages
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param message body forms.MessageForm true "Message"
// @Success 200 {object} forms.MessageOut "Message"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/ws [get]
func (wm *WebSocketManager) HandleMessages(conn *websocket.Conn, user *models.User) {
    defer conn.Close()

    for {
        var messageForm forms.MessageForm
        _, msg, err := conn.ReadMessage()
        if err != nil {
            logger.Error(context.Background(), fmt.Sprintf("Error while reading message from user %v: %v",
                user, err))
            return
        }
        conn, exists := wm.Connections[user.Id]
        if !exists {
            logger.Error(context.Background(), fmt.Sprintf("User %v not connected", user))
            return
        }

        err = json.Unmarshal(msg, &messageForm)
        if err != nil {
            logger.Error(context.Background(), "Error unmarshaling message:", err)
            writeErrorToWS(conn.Conn, fmt.Sprintf("Invalid message format: %v", err))
            continue
        }

        actionStruct := struct {
            Action string `json:"action"`
        }{}
        err = json.Unmarshal(msg, &actionStruct)
        if err != nil {
            logger.Error(context.Background(), "Error unmarshaling action:", err)
            writeErrorToWS(conn.Conn, fmt.Sprintf("Invalid action format: %v", err))
            continue
        }

        messageForm.SenderId = user.Id
        if messageForm.ChatId == uuid.Nil && messageForm.ReceiverId == uuid.Nil {
            logger.Error(context.Background(), "ChatId and ReceiverId cannot be both nil")
            writeErrorToWS(conn.Conn, "ChatId and ReceiverId cannot be both nil")
            continue
        }
        message := messageForm.ToMessageModel()

        message.ChatID, err = wm.MessageUseCase.SaveMessage(context.Background(), message)
        if err != nil {
            log.Println("Failed to save message:", err)
            writeErrorToWS(conn.Conn, fmt.Sprintf("Failed to save message: %v", err))
            continue
        }

        err = wm.SendMessageToChat(context.Background(), message.ChatID, message)
        if err != nil {
            log.Println("Failed to send message to chat:", err)
            writeErrorToWS(conn.Conn, fmt.Sprintf("Failed to send message to chat: %v", err))
            continue
        }
    }
}

func writeErrorToWS(conn *websocket.Conn, errMsg string) {
    err := conn.WriteJSON(forms.ErrorForm{
        Error: errMsg,
    })
    if err != nil {
        log.Println("Failed to send error message:", err)
    }
}
