package middleware

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/config/cors"
	"quickflow/internal/delivery/http/mocks"
	"quickflow/internal/models"
)

func TestContentTypeMiddleware(t *testing.T) {
	allowedTypes := []string{"application/json"}
	middlewareFunc := ContentTypeMiddleware(allowedTypes...)

	handler := middlewareFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Valid Content-Type", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid Content-Type", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	})
}

func TestCORSMiddleware(t *testing.T) {
	corsConfig := cors.CORSConfig{
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}
	middlewareFunc := CORSMiddleware(corsConfig)

	handler := middlewareFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Valid CORS Request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS Request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		r.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestSessionMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthUseCase(ctrl)
	middlewareFunc := SessionMiddleware(mockAuthService)

	handler := middlewareFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value("user").(models.User)
		assert.NotNil(t, user)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Valid Session", func(t *testing.T) {
		sessionID := uuid.New()
		user := models.User{Id: uuid.New()}

		mockAuthService.EXPECT().LookupUserSession(gomock.Any(), models.Session{SessionId: sessionID}).Return(user, nil).Times(1)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: sessionID.String()})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Session Cookie", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Session ID", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: "invalid-uuid"})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failed Authorization", func(t *testing.T) {
		sessionID := uuid.New()

		mockAuthService.EXPECT().LookupUserSession(gomock.Any(), models.Session{SessionId: sessionID}).Return(models.User{}, errors.New("auth error")).Times(1)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: sessionID.String()})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCSRFMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setCookie      bool
		setHeader      bool
		cookieValue    string
		headerValue    string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			setCookie:      false,
			setHeader:      false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST missing CSRF token",
			method:         http.MethodPost,
			setCookie:      false,
			setHeader:      false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST missing header",
			method:         http.MethodPost,
			setCookie:      true,
			cookieValue:    "123",
			setHeader:      false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST token mismatch",
			method:         http.MethodPost,
			setCookie:      true,
			cookieValue:    "123",
			setHeader:      true,
			headerValue:    "456",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST valid CSRF token",
			method:         http.MethodPost,
			setCookie:      true,
			cookieValue:    "123",
			setHeader:      true,
			headerValue:    "123",
			expectedStatus: http.StatusOK,
		},
	}

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)

			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "csrf_token",
					Value: tt.cookieValue,
				})
			}

			if tt.setHeader {
				req.Header.Set("X-CSRF-Token", tt.headerValue)
			}

			rr := httptest.NewRecorder()

			handler := CSRFMiddleware(okHandler)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
