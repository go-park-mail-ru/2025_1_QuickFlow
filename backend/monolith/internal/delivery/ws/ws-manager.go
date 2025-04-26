package ws

import (
	"context"
	"quickflow/monolith/internal/delivery/forms"
	models2 "quickflow/monolith/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IWebSocketManager interface {
	SendMessageToUser(ctx context.Context, userId uuid.UUID, message forms.MessageOut) error
	SendMessageToChat(ctx context.Context, message models2.Message, publicSenderInfo models2.PublicUserInfo, chatParticipants []models2.User) error
	IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
	HandlePing(conn *websocket.Conn)
}
