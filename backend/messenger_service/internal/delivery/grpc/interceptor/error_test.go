package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	message_errors "quickflow/messenger_service/internal/errors" // Путь к вашим ошибкам
)

func TestErrorInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedCode   codes.Code
		expectedMsg    string
		expectedReason string
	}{
		{
			name:           "ErrNotFound",
			err:            message_errors.ErrNotFound,
			expectedCode:   codes.NotFound,
			expectedMsg:    message_errors.ErrNotFound.Error(),
			expectedReason: "NOT_FOUND",
		},
		{
			name:           "ErrNotParticipant",
			err:            message_errors.ErrNotParticipant,
			expectedCode:   codes.PermissionDenied,
			expectedMsg:    message_errors.ErrNotParticipant.Error(),
			expectedReason: "NOT_PARTICIPANT",
		},
		{
			name:           "ErrInvalidChatCreationInfo",
			err:            message_errors.ErrInvalidChatCreationInfo,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    message_errors.ErrInvalidChatCreationInfo.Error(),
			expectedReason: "INVALID_CHAT_CREATION_INFO",
		},
		{
			name:           "ErrAlreadyInChat",
			err:            message_errors.ErrAlreadyInChat,
			expectedCode:   codes.AlreadyExists,
			expectedMsg:    message_errors.ErrAlreadyInChat.Error(),
			expectedReason: "ALREADY_IN_CHAT",
		},
		{
			name:           "ErrInvalidChatType",
			err:            message_errors.ErrInvalidChatType,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    message_errors.ErrInvalidChatType.Error(),
			expectedReason: "INVALID_CHAT_TYPE",
		},
		{
			name:           "ErrInvalidNumMessages",
			err:            message_errors.ErrInvalidNumMessages,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    message_errors.ErrInvalidNumMessages.Error(),
			expectedReason: "INVALID_NUM_MESSAGES",
		},
		{
			name:           "DefaultError",
			err:            errors.New("some other error"),
			expectedCode:   codes.Internal,
			expectedMsg:    "some other error",
			expectedReason: "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Симулируем обработку в контексте gRPC
			interceptor := ErrorInterceptor

			// Создаем mock handler, который просто возвращает ошибку
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, tt.err
			}

			// Вызываем interceptor
			_, err := interceptor(context.Background(), nil, nil, handler)

			// Проверяем, что ошибка была обработана корректно
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected error to be a grpc status error, got: %v", err)
				}

				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Equal(t, tt.expectedMsg, st.Message())

				// Проверяем детали ошибки
				var details *errdetails.ErrorInfo
				for _, d := range st.Details() {
					if e, ok := d.(*errdetails.ErrorInfo); ok {
						details = e
					}
				}
				if details != nil {
					assert.Equal(t, tt.expectedReason, details.Reason)
				} else {
					t.Fatalf("expected error to contain details, but got none")
				}
			} else {
				t.Fatalf("expected error, but got nil")
			}
		})
	}
}
