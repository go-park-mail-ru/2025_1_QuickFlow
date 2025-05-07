package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	community_errors "quickflow/community_service/internal/errors"
)

func TestErrorInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		inputError     error
		expectedCode   codes.Code
		expectedMsg    string
		expectedReason string
	}{
		{
			name:           "ErrorCommunityNameTooShort",
			inputError:     community_errors.ErrorCommunityNameTooShort,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "community name is too short",
			expectedReason: "COMMUNITY_NAME_TOO_SHORT",
		},
		{
			name:           "ErrorCommunityDescriptionTooLong",
			inputError:     community_errors.ErrorCommunityDescriptionTooLong,
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "community description is too long",
			expectedReason: "COMMUNITY_DESCRIPTION_TOO_LONG",
		},
		{
			name:           "ErrAlreadyExists",
			inputError:     community_errors.ErrAlreadyExists,
			expectedCode:   codes.AlreadyExists,
			expectedMsg:    "community with this name already exists",
			expectedReason: "COMMUNITY_ALREADY_EXISTS",
		},
		{
			name:           "ErrNotFound",
			inputError:     community_errors.ErrNotFound,
			expectedCode:   codes.NotFound,
			expectedMsg:    "not found",
			expectedReason: "COMMUNITY_NOT_FOUND",
		},
		{
			name:           "ErrForbidden",
			inputError:     community_errors.ErrForbidden,
			expectedCode:   codes.PermissionDenied,
			expectedMsg:    "forbidden",
			expectedReason: "FORBIDDEN",
		},
		{
			name:           "Default case",
			inputError:     errors.New("unknown error"),
			expectedCode:   codes.Internal,
			expectedMsg:    "unknown error",
			expectedReason: "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := ErrorInterceptor
			ctx := context.Background()
			req := &struct{}{}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, tt.inputError
			}

			_, err := interceptor(ctx, req, nil, handler)
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
