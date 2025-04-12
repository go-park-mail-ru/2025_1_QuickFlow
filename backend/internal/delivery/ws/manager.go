package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
)

type WebSocketManager struct {
	Connections map[uuid.UUID]*websocket.Conn
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Connections: make(map[uuid.UUID]*websocket.Conn),
	}
}

// AddConnection adds a new user connection to the manager
func (wm *WebSocketManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
	wm.Connections[userId] = conn
}

// RemoveConnection removes a user connection from the manager
func (wm *WebSocketManager) RemoveConnection(userId uuid.UUID) {
	if conn, exists := wm.Connections[userId]; exists {
		conn.Close()
		delete(wm.Connections, userId)
	}
}

// SendMessageToUser sends a message to a specific user
func (wm *WebSocketManager) SendMessageToUser(_ context.Context, userId uuid.UUID, message forms.MessageOut) error {
	conn, exists := wm.Connections[userId]
	if !exists {
		return fmt.Errorf("user not connected")
	}

	msgJSON, err := json.Marshal(message)
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
	conn, exists := wm.Connections[userId]
	return conn, exists
}
