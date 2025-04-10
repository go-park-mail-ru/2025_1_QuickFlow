package ws

//
//import (
//	"context"
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/golang/mock/gomock"
//	"github.com/google/uuid"
//	"github.com/gorilla/websocket"
//	"quickflow/internal/delivery/forms"
//	"quickflow/internal/models"
//	"quickflow/internal/usecase/mocks"
//)
//
//func TestMessageHandler(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockUseCase := mocks.NewMockIMessageRepository(ctrl)
//	mockWSManager := NewMockWS(ctrl)
//	h := NewMessageHandler(mockUseCase, mockWSManager)
//
//	tests := []struct {
//		name        string
//		user        models.User
//		chatIDs     []uuid.UUID
//		msg         forms.MessageForm
//		mockSetup   func()
//		expectError bool
//	}{
//		{
//			name: "Valid message",
//			user: models.User{Id: uuid.New()},
//			chatIDs: []uuid.UUID{
//				uuid.New(),
//			},
//			msg: forms.MessageForm{
//				ChatID: uuid.New().String(),
//				Text:   "Hello, world!",
//			},
//			mockSetup: func() {
//				mockUseCase.EXPECT().Send(gomock.Any()).Return(uuid.New(), nil)
//				mockWSManager.EXPECT().BroadcastMessage(gomock.Any(), gomock.Any())
//			},
//			expectError: false,
//		},
//		{
//			name: "Invalid chat ID",
//			user: models.User{Id: uuid.New()},
//			chatIDs: []uuid.UUID{
//				uuid.New(),
//			},
//			msg: forms.MessageForm{
//				ChatID: "invalid-uuid",
//				Text:   "Hello!",
//			},
//			mockSetup:   func() {},
//			expectError: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				r = r.WithContext(context.WithValue(r.Context(), "user", tt.user))
//
//				conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
//				if err != nil {
//					t.Fatalf("failed to upgrade websocket: %v", err)
//				}
//
//				defer conn.Close()
//
//				r = r.WithContext(context.WithValue(r.Context(), "wsConn", conn))
//
//				// Запускаем моковые ожидания
//				tt.mockSetup()
//
//				// Запускаем обработчик
//				h.MessageHandler(w, r)
//
//				// Отправляем начальное сообщение
//				initMsg, _ := json.Marshal(map[string]interface{}{
//					"chat_ids": tt.chatIDs,
//				})
//				conn.WriteMessage(websocket.TextMessage, initMsg)
//
//				// Отправляем тестовое сообщение
//				msg, _ := json.Marshal(tt.msg)
//				conn.WriteMessage(websocket.TextMessage, msg)
//			}))
//			defer server.Close()
//
//			// Подключаемся к веб-сокету
//			wsURL := "ws" + server.URL[4:]
//			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
//			if tt.expectError {
//				if err == nil {
//					t.Error("expected error but got none")
//				}
//			} else {
//				if err != nil {
//					t.Errorf("unexpected error: %v", err)
//				}
//			}
//		})
//	}
//}
