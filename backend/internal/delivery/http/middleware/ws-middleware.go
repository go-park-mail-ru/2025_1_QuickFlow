package middleware

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	http2 "quickflow/internal/delivery/http"
	"quickflow/internal/models"
	"quickflow/pkg/logger"
	httpUtils "quickflow/utils/http"
)

// WebSocketMiddleware устанавливает ws соединение с клиентом и обрабатывает сессии.
func WebSocketMiddleware(useCase http2.AuthUseCase) func(next http.Handler) http.Handler {
	// Upgrader можно создать один раз для всего приложения.
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// В данном примере разрешены все источники.
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
				httpUtils.WriteJSONError(w, "Failed to upgrade to WebSocket", http.StatusBadRequest)
				return
			}

			logger.Info(context.Background(), "[MIDDLEWARE] WebSocket connection established")

			// Устанавливаем WebSocket соединение и пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), "wsConn", conn)
			ctx = context.WithValue(ctx, "user", user)
			r = r.WithContext(ctx)

			// Передаем управление следующему обработчику
			defer func() {
				logger.Info(context.Background(), "[MIDDLEWARE] Closing WebSocket connection")
				conn.Close() // Закрываем соединение после завершения запроса
			}()
			next.ServeHTTP(w, r)
		})
	}
}
