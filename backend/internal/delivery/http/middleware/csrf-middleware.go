package middleware

import (
	"net/http"
)

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodDelete || r.Method == http.MethodPatch {
			csrfCookie, err := r.Cookie("csrf_token")
			if err != nil {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfHeader == "" || csrfCookie.Value != csrfHeader {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
