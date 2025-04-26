package http_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	http2 "quickflow/monolith/internal/delivery/http"
	"quickflow/monolith/internal/delivery/http/mocks"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/internal/usecase"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"
)

func TestMessageHandler_GetMessagesForChat(t *testing.T) {
	testTime := time.Now()
	userID := uuid.New()
	chatID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		chatID             string
		queryParams        map[string]string
		mockBehavior       func(mockMessageUC *mocks.MockMessageUseCase, mockProfileUC *mocks.MockProfileUseCase)
		expectedStatusCode int
	}{
		{
			name:    "OK",
			ctxUser: &models.User{Id: userID, Username: "testuser"},
			chatID:  chatID.String(),
			queryParams: map[string]string{
				"messages_count": "10",
				"ts":             testTime.Format("2006-01-02T15:04:05Z07:00"),
			},
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, mockProfileUC *mocks.MockProfileUseCase) {
				// Возвращаем тестовое сообщение с известным senderID
				testSenderID := uuid.New()
				mockMessageUC.EXPECT().
					GetMessagesForChat(gomock.Any(), chatID, userID, 10, gomock.Any()).
					Return([]models.Message{
						{SenderID: testSenderID},
					}, nil)

				mockProfileUC.EXPECT().
					GetPublicUsersInfo(gomock.Any(), []uuid.UUID{testSenderID}).
					Return(map[uuid.UUID]models.PublicUserInfo{
						testSenderID: {Id: testSenderID},
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:    "Invalid chat ID",
			ctxUser: &models.User{Id: userID, Username: "testuser"},
			chatID:  "invalid",
			queryParams: map[string]string{
				"messages_count": "10",
			},
			mockBehavior:       func(_ *mocks.MockMessageUseCase, _ *mocks.MockProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:    "Not participant",
			ctxUser: &models.User{Id: userID, Username: "testuser"},
			chatID:  chatID.String(),
			queryParams: map[string]string{
				"messages_count": "10",
			},
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, _ *mocks.MockProfileUseCase) {
				mockMessageUC.EXPECT().
					GetMessagesForChat(gomock.Any(), chatID, userID, 10, gomock.Any()).
					Return(nil, usecase.ErrNotParticipant)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:    "Invalid num messages",
			ctxUser: &models.User{Id: userID, Username: "testuser"},
			chatID:  chatID.String(),
			queryParams: map[string]string{
				"messages_count": "0",
			},
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, _ *mocks.MockProfileUseCase) {
				mockMessageUC.EXPECT().
					GetMessagesForChat(gomock.Any(), chatID, userID, 0, gomock.Any()).
					Return(nil, usecase.ErrInvalidNumMessages)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:    "Internal error",
			ctxUser: &models.User{Id: userID, Username: "testuser"},
			chatID:  chatID.String(),
			queryParams: map[string]string{
				"messages_count": "10",
			},
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, _ *mocks.MockProfileUseCase) {
				mockMessageUC.EXPECT().
					GetMessagesForChat(gomock.Any(), chatID, userID, 10, gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMessageUC := mocks.NewMockMessageUseCase(ctrl)
			mockAuthUC := mocks.NewMockAuthUseCase(ctrl)
			mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
			policy := bluemonday.NewPolicy()

			handler := http2.NewMessageHandler(mockMessageUC, mockAuthUC, mockProfileUC, policy)
			tc.mockBehavior(mockMessageUC, mockProfileUC)

			req := httptest.NewRequest(http.MethodGet, "/api/chats/"+tc.chatID+"/messages", nil)
			q := req.URL.Query()
			for k, v := range tc.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/api/chats/{chat_id}/messages", handler.GetMessagesForChat)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestMessageHandler_SendMessageToUsername(t *testing.T) {
	userID := uuid.New()
	recipientID := uuid.New()
	username := "recipient"
	messageText := "Hello, world!"

	testCases := []struct {
		name               string
		ctxUser            *models.User
		username           string
		requestBody        string
		mockBehavior       func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase)
		expectedStatusCode int
	}{
		{
			name:     "OK",
			ctxUser:  &models.User{Id: userID, Username: "sender"},
			username: username,
			requestBody: fmt.Sprintf(`{
				"text": "%s",
				"attachment_urls": []
			}`, messageText),
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase) {
				mockAuthUC.EXPECT().
					GetUserByUsername(gomock.Any(), username).
					Return(models.User{Id: recipientID}, nil)
				mockMessageUC.EXPECT().
					SaveMessage(gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Empty username",
			ctxUser:            &models.User{Id: userID, Username: "sender"},
			username:           "",
			requestBody:        `{}`,
			mockBehavior:       func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase) {},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "User not found",
			ctxUser:  &models.User{Id: userID, Username: "sender"},
			username: username,
			requestBody: fmt.Sprintf(`{
				"text": "%s"
			}`, messageText),
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase) {
				mockAuthUC.EXPECT().
					GetUserByUsername(gomock.Any(), username).
					Return(models.User{}, usecase.ErrNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:        "Invalid JSON",
			ctxUser:     &models.User{Id: userID, Username: "sender"},
			username:    username,
			requestBody: `{`,
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase) {
				mockAuthUC.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).Return(models.User{}, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:     "Save message error",
			ctxUser:  &models.User{Id: userID, Username: "sender"},
			username: username,
			requestBody: fmt.Sprintf(`{
				"text": "%s"
			}`, messageText),
			mockBehavior: func(mockMessageUC *mocks.MockMessageUseCase, mockAuthUC *mocks.MockAuthUseCase) {
				mockAuthUC.EXPECT().
					GetUserByUsername(gomock.Any(), username).
					Return(models.User{Id: recipientID}, nil)
				mockMessageUC.EXPECT().
					SaveMessage(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("save error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMessageUC := mocks.NewMockMessageUseCase(ctrl)
			mockAuthUC := mocks.NewMockAuthUseCase(ctrl)
			mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
			policy := bluemonday.NewPolicy()

			handler := http2.NewMessageHandler(mockMessageUC, mockAuthUC, mockProfileUC, policy)

			tc.mockBehavior(mockMessageUC, mockAuthUC)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/messages/"+tc.username,
				bytes.NewBufferString(tc.requestBody),
			)

			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/messages/{username}", handler.SendMessageToUsername)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
