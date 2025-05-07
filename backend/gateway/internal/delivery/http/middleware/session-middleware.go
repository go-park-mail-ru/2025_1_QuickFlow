package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	http2 "quickflow/gateway/internal/delivery/http"
	errors2 "quickflow/gateway/internal/errors"
	httpUtils "quickflow/gateway/utils/http"
	"quickflow/shared/models"
)

func SessionMiddleware(authUseCase http2.AuthUseCase) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if session exists
			session, err := r.Cookie("session")
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "invalid cookie", http.StatusUnauthorized)
				return
			}

			// parse session
			sessionUuid, err := uuid.Parse(session.Value)
			if err != nil {
				httpUtils.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse session", http.StatusBadRequest))
				return
			}

			// lookup user by session
			user, err := authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
			if err != nil {
				http.Error(w, "invalid cookie", http.StatusUnauthorized)
				return
			}

			// add user to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user", user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
