package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"quickflow/config"
	"quickflow/internal/models"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) LookupUserSession(ctx context.Context, session models.Session) (models.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockAuthService) CreateUser(ctx context.Context, user models.User) (uuid.UUID, models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAuthService) GetUser(ctx context.Context, authData models.LoginData) (models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAuthService) DeleteUserSession(ctx context.Context, session string) error {
	//TODO implement me
	panic("implement me")
}

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
	corsConfig := config.CORSConfig{
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
	mockAuthService := new(MockAuthService)
	middlewareFunc := SessionMiddleware(mockAuthService)

	handler := middlewareFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value("user").(models.User)
		assert.NotNil(t, user)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Valid Session", func(t *testing.T) {
		sessionID := uuid.New()
		user := models.User{Id: uuid.New()}
		mockAuthService.On("LookupUserSession", mock.Anything, models.Session{SessionId: sessionID}).Return(user, nil).Once()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: sessionID.String()})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)
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
		mockAuthService.On("LookupUserSession", mock.Anything, models.Session{SessionId: sessionID}).Return(models.User{}, errors.New("auth error")).Once()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: sessionID.String()})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}
