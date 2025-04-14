package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/pkg/logger"
	http2 "quickflow/utils/http"
)

type IWebSocketManager interface {
	AddConnection(userId uuid.UUID, conn *websocket.Conn)
	RemoveConnection(userId uuid.UUID)
	SendMessageToChat(ctx context.Context, message models.Message, publicSenderInfo models.PublicUserInfo,
		chatParticipants []models.User) error
	IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
	HandlePing(conn *websocket.Conn)
}

type MessageHandlerWS struct {
	MessageUseCase   MessageUseCase
	ChatUseCase      ChatUseCase
	profileUseCase   ProfileUseCase
	WebSocketManager IWebSocketManager
}

func NewMessageHandlerWS(messageUseCase MessageUseCase, chatUseCase ChatUseCase, profileUseCase ProfileUseCase, webSocketManager IWebSocketManager) *MessageHandlerWS {
	return &MessageHandlerWS{
		MessageUseCase:   messageUseCase,
		ChatUseCase:      chatUseCase,
		profileUseCase:   profileUseCase,
		WebSocketManager: webSocketManager,
	}
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
func (m *MessageHandlerWS) HandleMessages(_ http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())
	// extracting user from context
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while handling messages")
		return
	}

	conn, found := m.WebSocketManager.IsConnected(user.Id)
	if !found {
		logger.Error(ctx, "Failed to get WebSocket connection for user:", user)
		return
	}

	defer func() {
		m.profileUseCase.UpdateLastSeen(ctx, user.Id)
	}()

	for {
		var messageRequest forms.MessageRequest
		_, msg, err := conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				logger.Info(ctx, fmt.Sprintf("Connection closed normally by user %v: %v", user, closeErr))
			} else {
				logger.Error(ctx, fmt.Sprintf("Error reading message from user %v: %v", user, err))
			}

			return
		}

		conn, exists := m.WebSocketManager.IsConnected(user.Id)
		if !exists {
			logger.Error(ctx, fmt.Sprintf("User %v not connected", user))
			return
		}

		err = json.Unmarshal(msg, &messageRequest)
		if err != nil {
			logger.Error(ctx, "Error unmarshaling message:", err)
			writeErrorToWS(conn, fmt.Sprintf("Invalid message format: %v", err))
			continue
		}
		messageForm := messageRequest.Payload

		messageForm.SenderId = user.Id
		if messageForm.ChatId == uuid.Nil && messageForm.ReceiverId == uuid.Nil {
			logger.Error(ctx, "ChatId and ReceiverId cannot be both nil")
			writeErrorToWS(conn, "ChatId and ReceiverId cannot be both nil")
			continue
		}
		message := messageForm.ToMessageModel()

		message.ChatID, err = m.MessageUseCase.SaveMessage(ctx, message)
		if err != nil {
			log.Println("Failed to save message:", err)
			writeErrorToWS(conn, fmt.Sprintf("Failed to save message: %v", err))
			continue
		}

		// retrieving info to send message to all chat users
		publicSenderInfo, err := m.profileUseCase.GetPublicUserInfo(ctx, user.Id)
		if err != nil {
			log.Println("Failed to get public sender info:", err)
			writeErrorToWS(conn, fmt.Sprintf("Failed to get public sender info: %v", err))
			continue
		}
		chatParticipants, err := m.ChatUseCase.GetChatParticipants(ctx, message.ChatID)
		if err != nil {
			log.Println("Failed to get chat participants:", err)
			writeErrorToWS(conn, fmt.Sprintf("Failed to get chat participants: %v", err))
			continue
		}
		err = m.WebSocketManager.SendMessageToChat(ctx, message, publicSenderInfo, chatParticipants)
		if err != nil {
			log.Println("Failed to send message to chat:", err)
			writeErrorToWS(conn, fmt.Sprintf("Failed to send message to chat: %v", err))
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
