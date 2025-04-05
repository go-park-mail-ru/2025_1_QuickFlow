package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/internal/delivery/http/mocks"
	"quickflow/internal/models"
)

func TestFeedHandler_AddPost_TableDriven(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	user := models.User{Id: userID}

	tests := []struct {
		name           string
		inputPostForm  string
		contextUser    *models.User
		mockError      error
		expectedStatus int
		expectedPost   models.Post
	}{
		{
			name: "Successful post addition",
			inputPostForm: `{
				"desc": "Hello, world!",
				"pics": ["pic1.jpg", "pic2.jpg"]
			}`,
			contextUser:    &user,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedPost: models.Post{
				CreatorId: userID,
				Desc:      "Hello, world!",
				Images:    []string{"pic1.jpg", "pic2.jpg"},
				CreatedAt: time.Now(),
			},
		},
		{
			name: "Failed to add post (internal error)",
			inputPostForm: `{
				"desc": "This should fail",
				"pics": ["pic3.jpg"]
			}`,
			contextUser:    &user,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedPost: models.Post{
				CreatorId: userID,
				Desc:      "This should fail",
				Images:    []string{"pic3.jpg"},
				CreatedAt: time.Time{},
			},
		},
		{
			name:           "Invalid JSON request",
			inputPostForm:  `invalid json`,
			contextUser:    &user,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedPost:   models.Post{},
		},
		{
			name: "Missing user in context",
			inputPostForm: `{
				"desc": "Should fail due to missing user"
			}`,
			contextUser:    nil,
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
			expectedPost:   models.Post{},
		},
	}

	mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPostUseCase := mocks.NewMockPostUseCase(ctrl)
			handler := NewFeedHandler(mockPostUseCase, mockAuthUseCase)

			mockPostUseCase.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(tt.mockError).AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/post", bytes.NewBufferString(tt.inputPostForm))

			// Устанавливаем пользователя в контексте, если он есть
			if tt.contextUser != nil {
				req = req.WithContext(context.WithValue(req.Context(), "user", *tt.contextUser))
			}

			w := httptest.NewRecorder()

			handler.AddPost(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
