package middleware

import (
	"errors"
	"net/http"

	"quickflow/utils"
)

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfCookie, err := r.Cookie("csrf_token")
		var csrfToken string

		if errors.Is(err, http.ErrNoCookie) {
			csrfToken, _ = utils.GenerateCSRFToken()
		} else {
			csrfToken = csrfCookie.Value
		}

		if r.Method == http.MethodPost {
			clientToken := r.Header.Get("X-CSRF-Token")

			if clientToken != csrfToken {
				http.Error(w, "CSRF-токен недействителен", http.StatusForbidden)
				return
			}

			csrfToken, _ = utils.GenerateCSRFToken()
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		w.Header().Set("X-CSRF-Token", csrfToken)

		next.ServeHTTP(w, r)
	})
}
