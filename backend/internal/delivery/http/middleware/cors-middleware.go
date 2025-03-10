package middleware

import (
	"net/http"
	"quickflow/config"
	"strings"
)

// CORSMiddleware adds CORS headers to the response.
func CORSMiddleware(config config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Разрешенные домены
			allowedOrigins := map[string]bool{
				"https://www.quickflowapp.ru": true,
			}

			// Проверяем, не установлен ли уже CORS-заголовок
			if _, exists := w.Header()["Access-Control-Allow-Origin"]; !exists {
				if allowedOrigins[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}
			}

			// Если OPTIONS-запрос, отвечаем 204
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
