package http_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/gateway/internal/delivery/ws/mocks"
	"quickflow/shared/models"
)

func TestFriendsHandler_GetFriends(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)
	handler := http2.NewFriendHandler(mockFriendsUseCase, mockWS)

	userID := uuid.New()
	targetUserID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		queryParams        map[string]string
		mockBehavior       func()
		expectedStatusCode int
	}{
		{
			name:        "OK (Current User)",
			ctxUser:     &models.User{Id: userID, Username: "testuser"},
			queryParams: map[string]string{},
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					GetFriendsInfo(gomock.Any(), userID.String(), "", "").
					Return([]models.FriendInfo{}, false, 0, nil)
				mockWS.EXPECT().IsConnected(gomock.Any()).Return(nil, false).AnyTimes()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "OK (Specific User)",
			ctxUser:     &models.User{Id: userID, Username: "testuser"},
			queryParams: map[string]string{"user_id": targetUserID.String()},
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					GetFriendsInfo(gomock.Any(), targetUserID.String(), "", "").
					Return([]models.FriendInfo{}, false, 0, nil)
				mockWS.EXPECT().IsConnected(gomock.Any()).Return(nil, false).AnyTimes()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Error in UseCase",
			ctxUser:     &models.User{Id: userID, Username: "testuser"},
			queryParams: map[string]string{},
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					GetFriendsInfo(gomock.Any(), userID.String(), "", "").
					Return(nil, false, 0, errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "No User in Context",
			ctxUser:            nil,
			queryParams:        map[string]string{},
			mockBehavior:       func() {},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/api/friends", nil)
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
			handler.GetFriends(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestFriendsHandler_SendFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)
	handler := http2.NewFriendHandler(mockFriendsUseCase, mockWS)

	userID := uuid.New()
	receiverID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		inputBody          string
		mockBehavior       func()
		expectedStatusCode int
	}{
		{
			name:      "OK (Success)",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					IsExistsFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(false, nil)
				mockFriendsUseCase.EXPECT().
					SendFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Request Already Exists",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					IsExistsFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(true, nil)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:      "Bad JSON",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: `{"receiver_id": "invalid",`,
			mockBehavior: func() {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "Error in IsExistsFriendRequest",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					IsExistsFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(false, errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:      "Error in SendFriendRequest",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					IsExistsFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(false, nil)
				mockFriendsUseCase.EXPECT().
					SendFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodPost, "/api/friends", bytes.NewBufferString(tc.inputBody))
			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.SendFriendRequest(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestFriendsHandler_AcceptFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)
	handler := http2.NewFriendHandler(mockFriendsUseCase, mockWS)

	userID := uuid.New()
	receiverID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		inputBody          string
		mockBehavior       func()
		expectedStatusCode int
	}{
		{
			name:      "OK (Success)",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					AcceptFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Bad JSON",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: `{"receiver_id": "invalid",`,
			mockBehavior: func() {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "Error in AcceptFriendRequest",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"receiver_id":"%s"}`, receiverID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					AcceptFriendRequest(gomock.Any(), userID.String(), receiverID.String()).
					Return(errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodPost, "/api/friends/accept", bytes.NewBufferString(tc.inputBody))
			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.AcceptFriendRequest(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestFriendsHandler_DeleteFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)
	handler := http2.NewFriendHandler(mockFriendsUseCase, mockWS)

	userID := uuid.New()
	friendID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		inputBody          string
		mockBehavior       func()
		expectedStatusCode int
	}{
		{
			name:      "OK (Success)",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"friend_id":"%s"}`, friendID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					DeleteFriend(gomock.Any(), userID.String(), friendID.String()).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Bad JSON",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: `{"friend_id": "invalid",`,
			mockBehavior: func() {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "Error in DeleteFriend",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"friend_id":"%s"}`, friendID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					DeleteFriend(gomock.Any(), userID.String(), friendID.String()).
					Return(errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodDelete, "/api/friends", bytes.NewBufferString(tc.inputBody))
			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.DeleteFriend(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestFriendsHandler_Unfollow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockWS := mocks.NewMockIWebSocketManager(ctrl)
	handler := http2.NewFriendHandler(mockFriendsUseCase, mockWS)

	userID := uuid.New()
	friendID := uuid.New()

	testCases := []struct {
		name               string
		ctxUser            *models.User
		inputBody          string
		mockBehavior       func()
		expectedStatusCode int
	}{
		{
			name:      "OK (Success)",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"friend_id":"%s"}`, friendID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					Unfollow(gomock.Any(), userID.String(), friendID.String()).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Bad JSON",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: `{"friend_id": "invalid",`,
			mockBehavior: func() {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "Error in Unfollow",
			ctxUser:   &models.User{Id: userID, Username: "testuser"},
			inputBody: fmt.Sprintf(`{"friend_id":"%s"}`, friendID.String()),
			mockBehavior: func() {
				mockFriendsUseCase.EXPECT().
					Unfollow(gomock.Any(), userID.String(), friendID.String()).
					Return(errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodPost, "/api/friends/unfollow", bytes.NewBufferString(tc.inputBody))
			if tc.ctxUser != nil {
				ctx := context.WithValue(req.Context(), "user", *tc.ctxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.Unfollow(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
