package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/http/forms"
	forms2 "quickflow/gateway/internal/delivery/ws/forms"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type MessageService interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (*models.Message, error)
	GetMessagesForChat(ctx context.Context, chatId uuid.UUID, numMessages int, timestamp time.Time, userId uuid.UUID) ([]*models.Message, error)
	SendMessage(ctx context.Context, message *models.Message, userId uuid.UUID) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageId uuid.UUID) error
	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (time.Time, error)
	UpdateLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, timestamp time.Time, userAuthId uuid.UUID) error
}

type CommandHandler func(ctx context.Context, user models.User, payload json.RawMessage) error

// IWebSocketConnectionManager интерфейс для управления соединениями
type IWebSocketConnectionManager interface {
	AddConnection(userId uuid.UUID, conn *websocket.Conn)
	RemoveAndCloseConnection(userId uuid.UUID)
	IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
}

type IWebSocketRouter interface {
	RegisterHandler(command string, handler CommandHandler)
	Route(ctx context.Context, command string, user models.User, payload json.RawMessage) error
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

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while handling messages")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	conn, found := m.WebSocketManager.IsConnected(user.Id)
	if !found {
		logger.Error(ctx, fmt.Sprintf("WebSocket connection not found for user: %s", user.Id))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "WebSocket connection not found", http.StatusInternalServerError))
		return
	}

	defer func() {
		if err := m.profileUseCase.UpdateLastSeen(ctx, user.Id); err != nil {
			err = errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to update last seen: %s", err))
			http2.WriteJSONError(w, err)
		}
	}()

	for {
		var messageRequest forms2.MessageRequest

		_, msg, err := conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				logger.Info(ctx, fmt.Sprintf("WebSocket closed by user %v: %v", user.Id, closeErr))
			} else {
				logger.Error(ctx, fmt.Sprintf("Error reading WS message for user %v: %v", user.Id, err))
			}
			return
		}

		if err := json.Unmarshal(msg, &messageRequest); err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to unmarshal WS message: %v", err))
			writeErrorToWS(conn, fmt.Sprintf("Invalid message format: %v", err))
			continue
		}

		if err := m.WebSocketRouter.Route(ctx, messageRequest.Type, user, messageRequest.Payload); err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to route WS message: %v", err))
			writeErrorToWS(conn, fmt.Sprintf("Failed to process message: %v", err))
			continue
		}
	}
}

func writeErrorToWS(conn *websocket.Conn, errMsg string) {
	if err := conn.WriteJSON(forms.ErrorForm{ErrorCode: errMsg}); err != nil {
		log.Printf("Failed to send WS error message: %v", err)
	}
}
