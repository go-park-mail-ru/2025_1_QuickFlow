package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	feedback_errors "quickflow/feedback_service/internal/errors"
	"quickflow/shared/models"
)

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Invoke(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func TestErrorInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		inputError     error
		expectedCode   codes.Code
		expectedMsg    string
		expectedReason string
	}{
		{
			name:           "ErrNotFound",
			inputError:     feedback_errors.ErrNotFound,
			expectedCode:   codes.NotFound,
			expectedMsg:    "not found",
			expectedReason: "NOT_FOUND",
		},
		{
			name:           "ErrRespondent",
			inputError:     feedback_errors.ErrRespondent,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "invalid respondent",
			expectedReason: "INVALID_RESPONDENT",
		},
		{
			name:           "ErrRating",
			inputError:     feedback_errors.ErrRating,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "invalid rating",
			expectedReason: "INVALID_RATING",
		},
		{
			name:           "ErrTextTooLong",
			inputError:     feedback_errors.ErrTextTooLong,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "text is too long",
			expectedReason: "TEXT_TOO_LONG",
		},
		{
			name:           "Default case",
			inputError:     errors.New("some other error"),
			expectedCode:   codes.Internal,
			expectedMsg:    "some other error",
			expectedReason: "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler for the gRPC call
			mockHandler := new(mockHandler)
			mockHandler.On("Invoke", mock.Anything, mock.Anything).Return(nil, tt.inputError)

			// Use the ErrorInterceptor
			interceptor := ErrorInterceptor
			ctx := context.Background()
			req := &models.Feedback{} // Dummy feedback object

			// Execute the handler with the interceptor
			_, err := interceptor(ctx, req, nil, mockHandler.Invoke)

			if assert.Error(t, err) {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Equal(t, tt.expectedMsg, st.Message())
				assert.Contains(t, st.Message(), tt.expectedMsg)
			}
		})
	}
}
