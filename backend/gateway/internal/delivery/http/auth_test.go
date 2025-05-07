package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"

	"quickflow/gateway/internal/delivery/forms"
	http2 "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestAuthHandler_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockAuthUseCase(ctrl)
	handler := http2.NewAuthHandler(mockUC, bluemonday.UGCPolicy())

	type testCase struct {
		name               string
		inputBody          string
		mockBehavior       func(mockUC *mocks.MockAuthUseCase)
		expectedStatusCode int
		responseContains   string
	}

	testTable := []testCase{
		{
			name: "OK (Success)",
			inputBody: toJSON(forms.SignUpForm{
				Login:       "Timex",
				Password:    "228Amogus!",
				Name:        "John",
				Surname:     "Doe",
				Sex:         1,
				DateOfBirth: "2000-01-01",
			}),
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.New(), models.Session{
						SessionId:  uuid.New(),
						ExpireDate: time.Now().Add(time.Hour),
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			responseContains:   `"user_id"`,
		},
		{
			name:               "Bad JSON",
			inputBody:          `{"Username": "broken",`,
			mockBehavior:       func(mockUC *mocks.MockAuthUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Validation Error (no password)",
			inputBody: toJSON(forms.SignUpForm{
				Login: "testUser1",
			}),
			mockBehavior:       func(mockUC *mocks.MockAuthUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Internal Error from CreateUser",
			inputBody: toJSON(forms.SignUpForm{
				Login:       "Timex",
				Password:    "228Amogus!",
				Name:        "John",
				Surname:     "Doe",
				Sex:         1,
				DateOfBirth: "2000-01-01",
			}),
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.Nil, models.Session{}, errors.New("fail"))
			},
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockUC)

			req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBufferString(tc.inputBody))
			rr := httptest.NewRecorder()

			handler.SignUp(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code, "Статус код должен совпадать")

			if tc.responseContains != "" {
				assert.Contains(t, rr.Body.String(), tc.responseContains)
			}
		})
	}
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockAuthUseCase(ctrl)
	handler := http2.NewAuthHandler(mockUC, bluemonday.UGCPolicy())

	type testCase struct {
		name               string
		inputBody          string
		mockBehavior       func(mockUC *mocks.MockAuthUseCase)
		expectedStatusCode int
		responseContains   string
	}

	testTable := []testCase{
		{
			name: "OK (Success)",
			inputBody: toJSON(forms.AuthForm{
				Login:    "Timex",
				Password: "228Amogus!",
			}),
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					AuthUser(gomock.Any(), gomock.Any()).
					Return(models.Session{
						SessionId:  uuid.New(),
						ExpireDate: time.Now().Add(time.Hour),
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			responseContains:   "",
		},
		{
			name:               "Bad JSON",
			inputBody:          `{"Username":"broken",`,
			mockBehavior:       func(mockUC *mocks.MockAuthUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Unauthorized",
			inputBody: toJSON(forms.AuthForm{
				Login:    "wrong",
				Password: "wrong",
			}),
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					AuthUser(gomock.Any(), gomock.Any()).
					Return(models.Session{}, errors.New("fail"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockUC)

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(tc.inputBody))
			rr := httptest.NewRecorder()

			handler.Login(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			if tc.responseContains != "" {
				assert.Contains(t, rr.Body.String(), tc.responseContains)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockAuthUseCase(ctrl)
	handler := http2.NewAuthHandler(mockUC, bluemonday.UGCPolicy())

	sessionID := uuid.New().String()

	type testCase struct {
		name               string
		setCookie          bool
		mockBehavior       func(mockUC *mocks.MockAuthUseCase)
		expectedStatusCode int
	}

	testTable := []testCase{
		{
			name:      "OK (Success)",
			setCookie: true,
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					LookupUserSession(gomock.Any(), models.Session{SessionId: uuid.MustParse(sessionID)}).
					Return(models.User{}, nil)
				mockUC.EXPECT().
					DeleteUserSession(gomock.Any(), sessionID).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "No Cookie",
			setCookie:          false,
			mockBehavior:       func(mockUC *mocks.MockAuthUseCase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:      "Lookup Error",
			setCookie: true,
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					LookupUserSession(gomock.Any(), gomock.Any()).
					Return(models.User{}, errors.New("not found"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:      "Delete Error",
			setCookie: true,
			mockBehavior: func(mockUC *mocks.MockAuthUseCase) {
				mockUC.EXPECT().
					LookupUserSession(gomock.Any(), gomock.Any()).
					Return(models.User{}, nil)
				mockUC.EXPECT().
					DeleteUserSession(gomock.Any(), sessionID).
					Return(errors.New("fail"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockUC)

			req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
			if tc.setCookie {
				req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
			}
			rr := httptest.NewRecorder()

			handler.Logout(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
