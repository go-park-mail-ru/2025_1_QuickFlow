package ws

//import (
//	"context"
//	"github.com/google/uuid"
//	"github.com/gorilla/websocket"
//	"log"
//	"net/http"
//	"time"
//)
//
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//func HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, wsManager WebSocketManager) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println("Failed to upgrade connection:", err)
//		return
//	}
//	userId := uuid.New() // Пример: здесь может быть ID из токена пользователя
//	wsConn := &WebSocketConnection{
//		UserId:   userId,
//		Conn:     &Connection{ /* Здесь можно передавать реальное соединение WebSocket */ },
//		LastSeen: time.Now(),
//	}
//
//	err = wsManager.ConnectUser(userId, wsConn)
//	if err != nil {
//		log.Println("Error connecting user:", err)
//		return
//	}
//
//	// Обработка получения сообщений и отправки их
//	for {
//		_, message, err := conn.ReadMessage()
//		if err != nil {
//			log.Println("Error reading message:", err)
//			break
//		}
//
//		// Пример: обработка сообщения и рассылка
//		wsManager.SendMessageToUser(context.Background(), userId, string(message))
//	}
//
//	// Отключение пользователя по завершению соединения
//	wsManager.DisconnectUser(userId)
//}
