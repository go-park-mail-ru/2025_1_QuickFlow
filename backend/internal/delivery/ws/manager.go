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

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/pkg/logger"
)

type WebSocketManager struct {
	Connections map[uuid.UUID]*websocket.Conn
	mu          sync.RWMutex
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Connections: make(map[uuid.UUID]*websocket.Conn),
	}
}

// AddConnection adds a new user connection to the manager
func (wm *WebSocketManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
	wm.mu.Lock()
	wm.Connections[userId] = conn
	wm.mu.Unlock()
}

// RemoveAndCloseConnection removes a user connection from the manager and closes it
func (wm *WebSocketManager) RemoveAndCloseConnection(userId uuid.UUID) {
	wm.mu.Lock()
	if _, exists := wm.Connections[userId]; exists {
		delete(wm.Connections, userId)
	}
	wm.mu.Unlock()
}

// SendMessageToUser sends a message to a specific user
func (wm *WebSocketManager) SendMessageToUser(_ context.Context, userId uuid.UUID, message forms.MessageOut) error {
	wm.mu.RLock()
	conn, exists := wm.Connections[userId]
	wm.mu.RUnlock()

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
func (wm *WebSocketManager) SendMessageToChat(ctx context.Context, message models.Message, publicSenderInfo models.PublicUserInfo, chatParticipants []models.User) error {
	for _, user := range chatParticipants {
		err := wm.SendMessageToUser(ctx, user.Id, forms.ToMessageOut(message, publicSenderInfo))
		if err != nil {
			log.Println("Failed to send message to user:", user.Id, err)
		}
	}

	return nil
}

func (wm *WebSocketManager) IsConnected(userId uuid.UUID) (*websocket.Conn, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	conn, exists := wm.Connections[userId]
	return conn, exists
}

func (wm *WebSocketManager) HandlePing(conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		logger.Info(context.Background(), "Received pong:", appData)
		return nil
	})

	go func() {
		for {
			time.Sleep(30 * time.Second) // sending ping every 30 seconds
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Info(context.Background(), "Failed to send ping:", err)
				return
			}
		}
	}()
}
