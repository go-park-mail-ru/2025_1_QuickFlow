package ws

import (
	"context"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"quickflow/monolith/internal/delivery/forms"
	"quickflow/monolith/internal/models"
)

type IWebSocketManager interface {
	SendMessageToUser(ctx context.Context, userId uuid.UUID, message forms.MessageOut) error
	SendMessageToChat(ctx context.Context, message models.Message, publicSenderInfo models.PublicUserInfo, chatParticipants []models.User) error
	IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
	HandlePing(conn *websocket.Conn)
}
