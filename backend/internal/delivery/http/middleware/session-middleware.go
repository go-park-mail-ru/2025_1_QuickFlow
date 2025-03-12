package middleware

import (
	"context"
	"errors"
	"net/http"
	http2 "quickflow/internal/delivery/http"
	httpUtils "quickflow/utils/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"quickflow/internal/models"
)

func SessionMiddleware(authUseCase http2.AuthUseCase) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if session exists
			session, err := r.Cookie("session")
			if errors.Is(err, http.ErrNoCookie) {
				httpUtils.WriteJSONError(w, "Authorization needed", http.StatusUnauthorized)
				return
			}

			// parse session
			sessionUuid, err := uuid.Parse(session.Value)
			if err != nil {
				httpUtils.WriteJSONError(w, "Failed to parse session", http.StatusBadRequest)
				return
			}

			// lookup user by session
			user, err := authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
			if err != nil {
				httpUtils.WriteJSONError(w, "Failed to authorize user", http.StatusUnauthorized)
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
