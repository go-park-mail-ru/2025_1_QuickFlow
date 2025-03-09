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

            // Всегда добавляем заголовки CORS, даже если запрос ошибочный
            w.Header().Set("Vary", "Origin")
            if origin != "" {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            }
            w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
            w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

            if config.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }

            // Preflight-запрос
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)

            // Если ответ 404, добавляем CORS-заголовки вручную
            if w.Header().Get("Access-Control-Allow-Origin") == "" {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            }
        })
    }
}
