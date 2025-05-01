package middleware

import (
	"context"
	"fmt"
	"net/http"
	"quickflow/pkg/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(context.Background(), fmt.Sprintf("Panic: %v, URL: %s", err, r.URL.Path))
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 - Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
