package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func ContentTypeMiddleware(allowedTypes ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("Content-Type")
			for _, allowed := range allowedTypes {
				if strings.HasPrefix(contentType, allowed) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		})
	}
}
