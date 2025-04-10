package ws

//
//import (
//	"github.com/google/uuid"
//	"sync"
//	"testing"
//)
//
//type mockConn struct {
//	writeLock sync.Mutex
//	messages  []interface{}
//	closed    bool
//}
//
//func (c *mockConn) WriteJSON(v interface{}) error {
//	c.writeLock.Lock()
//	defer c.writeLock.Unlock()
//	c.messages = append(c.messages, v)
//	return nil
//}
//
//func (c *mockConn) Close() error {
//	c.closed = true
//	return nil
//}
//
//func TestWSManager(t *testing.T) {
//	tests := []struct {
//		name          string
//		action        func(m *WSManager, conn *mockConn, chatID uuid.UUID)
//		expectMsgSent bool
//		expectClosed  bool
//	}{
//		{
//			name: "Add client and broadcast message",
//			action: func(m *WSManager, conn *mockConn, chatID uuid.UUID) {
//				m.AddClient(conn, []uuid.UUID{chatID})
//				m.BroadcastMessage(chatID, "test message")
//			},
//			expectMsgSent: true,
//		},
//		{
//			name: "Remove client",
//			action: func(m *WSManager, conn *mockConn, chatID uuid.UUID) {
//				m.AddClient(conn, []uuid.UUID{chatID})
//				m.RemoveClient(conn)
//			},
//			expectClosed: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			manager := NewWSManager()
//			mockConn := &mockConn{}
//			chatID := uuid.New()
//
//			tt.action(manager, mockConn, chatID)
//
//			if tt.expectMsgSent {
//				if len(mockConn.messages) == 0 {
//					t.Errorf("expected message to be sent, but got none")
//				}
//			}
//			if tt.expectClosed && !mockConn.closed {
//				t.Errorf("expected connection to be closed, but it was open")
//			}
//		})
//	}
//}
