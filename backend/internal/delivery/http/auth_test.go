package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"quickflow/internal/models"
	http2 "quickflow/test/http"
)

func TestAuthHandler_SignUp(t *testing.T) {
	mockAuthUseCase := new(http2.MockAuthUseCase)
	handler := NewAuthHandler(mockAuthUseCase)

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
			mockAuthUseCase.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(tt.inputForm))
			w := httptest.NewRecorder()

			mockAuthUseCase.On("CreateUser", mock.Anything, mock.Anything).
				Return(userID, session, tt.mockError).
				Maybe()

			handler.SignUp(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockAuthUseCase := new(http2.MockAuthUseCase)
	handler := NewAuthHandler(mockAuthUseCase)

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
			mockAuthUseCase.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.inputJSON))
			w := httptest.NewRecorder()

			mockAuthUseCase.On("GetUser", mock.Anything, mock.Anything).Return(session, tt.mockError).Maybe()

			handler.Login(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	mockAuthUseCase := new(http2.MockAuthUseCase)
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
			mockAuthUseCase.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()

			mockAuthUseCase.On("LookupUserSession", mock.Anything, session).
				Return(models.User{}, tt.mockLookupErr).
				Maybe()

			mockAuthUseCase.On("DeleteUserSession", mock.Anything, sessionID.String()).
				Return(tt.mockDeleteErr).
				Maybe()

			handler.Logout(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockAuthUseCase.AssertExpectations(t)
		})
	}
}
