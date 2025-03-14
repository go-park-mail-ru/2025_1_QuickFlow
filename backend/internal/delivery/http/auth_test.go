package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"quickflow/internal/delivery/http/mocks"
	"quickflow/internal/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	session := models.Session{SessionId: uuid.New(), ExpireDate: time.Now().Add(time.Hour)}

	tests := []struct {
		name           string
		inputForm      string
		mockError      error
		expectedStatus int
	}{
		{
			name: "Successful sign-up",
			inputForm: `{
				"username": "Timex1",
				"firstname": "Matvey",
				"lastname": "Mitrofanov",
				"sex": 1,
				"birth_date": "2000.01.01",
				"password": "amoguS228!"
			}`,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			inputForm:      `invalid json`,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User already exists",
			inputForm: `{
				"username": "Timex1",
				"firstname": "Matvey",
				"lastname": "Mitrofanov",
				"sex": 1,
				"birth_date": "2000.01.01",
				"password": "amoguS228!"
			}`,
			mockError:      errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)
			handler := NewAuthHandler(mockAuthUseCase)

			mockAuthUseCase.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(userID, session, tt.mockError).AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(tt.inputForm))
			w := httptest.NewRecorder()

			handler.SignUp(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	session := models.CreateSession()

	tests := []struct {
		name           string
		inputJSON      string
		mockError      error
		expectedStatus int
	}{
		{
			name: "Successful login",
			inputJSON: `{
				"login": "testuser",
				"password": "password123"
			}`,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			inputJSON:      `invalid json`,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Unauthorized user",
			inputJSON: `{
				"login": "nonexistentuser",
				"password": "wrongpassword"
			}`,
			mockError:      errors.New("unauthorized"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)
			handler := NewAuthHandler(mockAuthUseCase)

			mockAuthUseCase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(session, tt.mockError).AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.Login(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)
	handler := NewAuthHandler(mockAuthUseCase)

	sessionID := uuid.New()
	session := models.Session{SessionId: sessionID}
	cookie := &http.Cookie{Name: "session", Value: sessionID.String()}

	tests := []struct {
		name           string
		mockLookupErr  error
		mockDeleteErr  error
		expectedStatus int
	}{
		{
			name:           "Successful logout",
			mockLookupErr:  nil,
			mockDeleteErr:  nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthUseCase.EXPECT().LookupUserSession(gomock.Any(), session).Return(models.User{}, tt.mockLookupErr).AnyTimes()
			mockAuthUseCase.EXPECT().DeleteUserSession(gomock.Any(), sessionID.String()).Return(tt.mockDeleteErr).AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()

			handler.Logout(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
