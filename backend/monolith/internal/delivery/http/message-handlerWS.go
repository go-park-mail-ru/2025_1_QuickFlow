package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"quickflow/monolith/internal/delivery/forms"
	forms2 "quickflow/monolith/internal/delivery/ws/forms"
	models2 "quickflow/monolith/internal/models"
	"quickflow/monolith/pkg/logger"
	http2 "quickflow/monolith/utils/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
)

type MessageUseCase interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (models2.Message, error)
	GetMessagesForChat(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, numMessages int, timestamp time.Time) ([]models2.Message, error)
	SaveMessage(ctx context.Context, message models2.Message) (uuid.UUID, error)
	DeleteMessage(ctx context.Context, messageId uuid.UUID) error
	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*time.Time, error)
	UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId uuid.UUID, userId uuid.UUID) error
}

type CommandHandler func(ctx context.Context, user models2.User, payload json.RawMessage) error

// IWebSocketConnectionManager интерфейс для управления соединениями
type IWebSocketConnectionManager interface {
	AddConnection(userId uuid.UUID, conn *websocket.Conn)
	RemoveAndCloseConnection(userId uuid.UUID)
	IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
}

type IWebSocketRouter interface {
	RegisterHandler(command string, handler CommandHandler)
	Route(ctx context.Context, command string, user models2.User, payload json.RawMessage) error
}

// MessageListenerWS Обработчик сообщений
type MessageListenerWS struct {
	profileUseCase   ProfileUseCase
	WebSocketManager IWebSocketConnectionManager
	WebSocketRouter  IWebSocketRouter
	policy           *bluemonday.Policy
}

func NewMessageListenerWS(profileUseCase ProfileUseCase, webSocketManager IWebSocketConnectionManager, webSocketRouter IWebSocketRouter, policy *bluemonday.Policy) *MessageListenerWS {
	return &MessageListenerWS{
		profileUseCase:   profileUseCase,
		WebSocketManager: webSocketManager,
		policy:           policy,
		WebSocketRouter:  webSocketRouter,
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
func (m *MessageListenerWS) HandleMessages(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())
	// Извлекаем пользователя из контекста
	user, ok := ctx.Value("user").(models2.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while handling messages")
		return
	}

	conn, found := m.WebSocketManager.IsConnected(user.Id)
	if !found {
		logger.Error(ctx, "Failed to get WebSocket connection for user:", user)
		return
	}

	// Завершаем работу по обновлению времени последнего посещения
	defer func() {
		if err := m.profileUseCase.UpdateLastSeen(ctx, user.Id); err != nil {
			http2.WriteJSONError(w, "Failed to update last seen", http.StatusBadRequest)
			return
		}
	}()

	for {
		var messageRequest forms2.MessageRequest
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

		// Десериализуем сообщение
		err = json.Unmarshal(msg, &messageRequest)
		if err != nil {
			logger.Error(ctx, "Error unmarshaling message:", err)
			writeErrorToWS(conn, fmt.Sprintf("Invalid message format: %v", err))
			continue
		}

		// Маршрутизируем команду через WebSocketRouter
		command := messageRequest.Type // предполагается, что в запросе будет команда, например, "message" или "ping"
		err = m.WebSocketRouter.Route(ctx, command, user, messageRequest.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Error handling message: %v", err))
			writeErrorToWS(conn, fmt.Sprintf("Failed to process message: %v", err))
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
