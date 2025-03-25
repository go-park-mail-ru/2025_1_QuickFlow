package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"quickflow/config/cors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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
