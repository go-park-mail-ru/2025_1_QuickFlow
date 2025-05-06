package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	http2 "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/ws"
	errors2 "quickflow/gateway/internal/errors"
	httpUtils "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

// WebSocketMiddleware устанавливает ws соединение с клиентом и обрабатывает сессии.
func WebSocketMiddleware(connManager http2.IWebSocketConnectionManager, handler ws.PingHandler) func(next http.Handler) http.Handler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(context.Background(), "[MIDDLEWARE] WebSocket request received on:", r.URL.Path)

			user, ok := r.Context().Value("user").(models.User)
			if !ok {
				logger.Error(r.Context(), "Failed to get user from context while upgrading to WebSocket")
				httpUtils.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
				return
			}

			// Апгрейд соединения на WebSocket
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Error(r.Context(), fmt.Sprintf("Failed to upgrade connection to WebSocket: %s", err.Error()))
				httpUtils.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to upgrade to WebSocket", http.StatusBadRequest))
				return
			}

			logger.Info(context.Background(), "[MIDDLEWARE] WebSocket connection established")

			// Устанавливаем WebSocket соединение и пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), "wsConn", conn)
			ctx = context.WithValue(ctx, "user", user)
			r = r.WithContext(ctx)

			connManager.AddConnection(user.Id, conn)

			// Обрабатываем ping/pong сообщения
			handler.Handle(ctx, conn)

			// Передаем управление следующему обработчику
			defer func() {
				logger.Info(context.Background(), "[MIDDLEWARE] Closing WebSocket connection")
				connManager.RemoveAndCloseConnection(user.Id)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
