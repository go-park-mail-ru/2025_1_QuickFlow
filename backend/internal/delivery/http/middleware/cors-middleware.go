package middleware

import (
    "log"
    "net/http"
    "quickflow/config"
    "strings"
)

// CORSMiddleware adds CORS headers to the response.
func CORSMiddleware(config config.CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Проверяем, входит ли Origin в список разрешенных
            allowed := false
            for _, o := range config.AllowedOrigins {
                if o == origin {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Vary", "Origin") // Важно для кэширования CORS
            }

            w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
            w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

            if len(config.ExposedHeaders) > 0 {
                w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
            }

            if config.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }

            if config.Debug {
                log.Printf("CORS Debug: %s %s - Origin: %s", r.Method, r.URL.Path, origin)
            }

            // Preflight-запрос (OPTIONS)
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
