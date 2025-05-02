package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	http2 "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/ws"
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
				logger.Error(r.Context(), "Failed to get user from context while adding post")
				httpUtils.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
				return
			}
			// Апгрейд соединения на WebSocket
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Info(r.Context(), "Failed to upgrade connection to WebSocket:", err)
				httpUtils.WriteJSONError(w, "Failed to upgrade to WebSocket", http.StatusBadRequest)
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
