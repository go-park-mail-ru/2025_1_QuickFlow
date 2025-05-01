package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/internal/delivery/forms"
	http2 "quickflow/internal/delivery/http"
	"quickflow/internal/delivery/http/mocks"
	"quickflow/internal/models"
	"quickflow/internal/usecase"
)

func TestChatHandler_GetUserChats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUC := mocks.NewMockChatUseCase(ctrl)
	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)

	handler := http2.NewChatHandler(mockChatUC, mockProfileUC, mockWS)

	type testCase struct {
		name               string
		ctxUser            *models.User
		queryParams        string
		mockBehavior       func()
		expectedStatusCode int
		validateResponse   func(t *testing.T, rr *httptest.ResponseRecorder)
	}

	myUserID := uuid.New()
	otherUser1 := uuid.New()
	otherUser2 := uuid.New()
	chatID1 := uuid.New()
	chatID2 := uuid.New()
	fakeNow := time.Now()

	testTable := []testCase{
		{
			name:               "No user in context",
			ctxUser:            nil,
			queryParams:        "chats_count=10",
			mockBehavior:       func() {},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Failed to get user from context")
			},
		},
		{
			name:               "Invalid query parameters",
			ctxUser:            &models.User{Id: myUserID, Username: "testuser"},
			queryParams:        "chats_count=notanumber",
			mockBehavior:       func() {},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Failed to parse query params")
			},
		},
		{
			name:        "User has no chats",
			ctxUser:     &models.User{Id: myUserID, Username: "testuser"},
			queryParams: "chats_count=10",
			mockBehavior: func() {
				mockChatUC.EXPECT().
					GetUserChats(gomock.Any(), myUserID).
					Return(nil, usecase.ErrNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "user has no chats")
			},
		},
		{
			name:        "Internal error in GetUserChats",
			ctxUser:     &models.User{Id: myUserID, Username: "testuser"},
			queryParams: "chats_count=10",
			mockBehavior: func() {
				mockChatUC.EXPECT().
					GetUserChats(gomock.Any(), myUserID).
					Return(nil, errors.New("db failure"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Failed to fetch chats")
			},
		},
		{
			name:        "Internal error in GetPublicUsersInfo",
			ctxUser:     &models.User{Id: myUserID, Username: "testuser"},
			queryParams: "chats_count=10",
			mockBehavior: func() {
				chat1 := models.Chat{
					ID:   chatID1,
					Type: models.ChatTypePrivate,
					LastMessage: models.Message{
						ID:       uuid.New(),
						SenderID: otherUser1,
					},
				}
				mockChatUC.EXPECT().
					GetUserChats(gomock.Any(), myUserID).
					Return([]models.Chat{chat1}, nil)

				mockProfileUC.EXPECT().
					GetPublicUsersInfo(gomock.Any(), []uuid.UUID{otherUser1}).
					Return(nil, errors.New("profile error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Failed to fetch last messages users info")
			},
		},
		{
			name:        "Success - two chats (branch with last message & branch with getOtherPrivateChatParticipant)",
			ctxUser:     &models.User{Id: myUserID, Username: "testuser"},
			queryParams: "chats_count=10",
			mockBehavior: func() {
				chat1 := models.Chat{
					ID:   chatID1,
					Type: models.ChatTypePrivate,
					LastMessage: models.Message{
						ID:       uuid.New(), // ненулевой ID
						SenderID: otherUser1,
					},
				}
				chat2 := models.Chat{
					ID:          chatID2,
					Type:        models.ChatTypePrivate,
					LastMessage: models.Message{}, // все поля — нули
				}
				mockChatUC.EXPECT().
					GetUserChats(gomock.Any(), myUserID).
					Return([]models.Chat{chat1, chat2}, nil)

				publicInfo1 := models.PublicUserInfo{
					Id:       otherUser1,
					Username: "other1",
					LastSeen: fakeNow.Add(-5 * time.Minute),
				}
				mockProfileUC.EXPECT().
					GetPublicUsersInfo(gomock.Any(), gomock.Eq([]uuid.UUID{otherUser1, uuid.Nil})).
					Return(map[uuid.UUID]models.PublicUserInfo{
						otherUser1: publicInfo1,
					}, nil)

				mockWS.EXPECT().
					IsConnected(otherUser1).
					Return(nil, true)

				mockChatUC.EXPECT().
					GetChatParticipants(gomock.Any(), chatID2).
					Return([]models.User{
						{Id: myUserID},
						{Id: otherUser2},
					}, nil)

				publicInfo2 := models.PublicUserInfo{
					Id:       otherUser2,
					Username: "other2",
					LastSeen: fakeNow.Add(-10 * time.Minute),
				}
				mockProfileUC.EXPECT().
					GetPublicUserInfo(gomock.Any(), otherUser2).
					Return(publicInfo2, nil)

				mockWS.EXPECT().
					IsConnected(otherUser2).
					Return(nil, true)
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var chatsOut []forms.ChatOut
				err := json.NewDecoder(rr.Body).Decode(&chatsOut)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(chatsOut))
			},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			baseURL := "/api/chats"
			queryStr := tc.queryParams
			u, _ := url.Parse(baseURL)
			u.RawQuery = queryStr

			req := httptest.NewRequest(http.MethodGet, u.String(), nil)

			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.GetUserChats(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code, "Статус код должен совпадать")

			if tc.validateResponse != nil {
				tc.validateResponse(t, rr)
			}
		})
	}
}
