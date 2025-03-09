package middleware

import (
	"log"
	"net/http"
	"quickflow/config"
	"strings"
)

// CORSMiddleware adds CORS headers to the response.
func CORSMiddleware(config config.CORSConfig) func(http.Handler) http.Handler {
	allowedOrigins := strings.Join(config.AllowedOrigins, ", ")
	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(config.ExposedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// setting CORS headers
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			// Exposing headers
			if exposedHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
			}

			// Enabling credentials if allowed
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.Debug {
				log.Printf("CORS debug: %s %s", r.Method, r.URL.Path)
			}

			// Handling preflight requests
			if r.Method == http.MethodOptions {
				if config.OptionsPassthrough {
					next.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
