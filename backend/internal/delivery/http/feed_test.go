package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"quickflow/config"
	"quickflow/internal/delivery/forms"
	http2 "quickflow/internal/delivery/http"
	"quickflow/internal/delivery/http/mocks"
	"quickflow/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeedHandler_GetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUseCase := mocks.NewMockPostUseCase(ctrl)
	mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)

	handler := http2.NewFeedHandler(mockPostUseCase, mockAuthUseCase)

	user := models.User{Id: uuid.New()}

	tests := []struct {
		name           string
		inputFeedForm  forms.FeedForm
		mockPosts      []models.Post
		mockError      error
		expectedStatus int
		expectedLen    int
		passUser       bool
	}{
		{
			name: "Successful fetch with 2 posts",
			inputFeedForm: forms.FeedForm{
				Posts: 2,
				Ts:    time.Now().Format(config.TimeStampLayout),
			},
			mockPosts: []models.Post{
				{CreatorId: uuid.New(), Desc: "Post 1"},
				{CreatorId: uuid.New(), Desc: "Post 2"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLen:    2,
			passUser:       true,
		},
		{
			name: "Empty feed",
			inputFeedForm: forms.FeedForm{
				Posts: 0,
				Ts:    time.Now().Format(config.TimeStampLayout),
			},
			mockPosts:      []models.Post{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLen:    0,
			passUser:       true,
		},
		{
			name: "FetchFeed returns error",
			inputFeedForm: forms.FeedForm{
				Posts: 3,
				Ts:    time.Now().Format(config.TimeStampLayout),
			},
			mockPosts:      nil,
			mockError:      errors.New("fetch error"),
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
			passUser:       true,
		},
		{
			name:           "Invalid JSON request",
			inputFeedForm:  forms.FeedForm{},
			mockPosts:      nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
			passUser:       true,
		},
		{
			name:           "No user in context passed",
			inputFeedForm:  forms.FeedForm{},
			mockPosts:      nil,
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
			passUser:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody bytes.Buffer
			if tt.name != "Invalid JSON request" {
				json.NewEncoder(&reqBody).Encode(tt.inputFeedForm)
			} else {
				reqBody.WriteString("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/feed", &reqBody)
			if tt.passUser {
				req = req.WithContext(context.WithValue(req.Context(), "user", user))
			}
			w := httptest.NewRecorder()

			if tt.passUser && tt.name != "Invalid JSON request" {
				mockPostUseCase.EXPECT().
					FetchFeed(gomock.Any(), user, tt.inputFeedForm.Posts, gomock.Any()).
					Return(tt.mockPosts, tt.mockError).
					AnyTimes()
			}

			handler.GetFeed(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var responsePosts []forms.PostOut
				json.NewDecoder(resp.Body).Decode(&responsePosts)
				assert.Len(t, responsePosts, tt.expectedLen)
			}
		})
	}
}
