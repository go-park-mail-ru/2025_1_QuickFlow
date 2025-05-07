package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	errors2 "quickflow/gateway/internal/errors"
	httpUtils "quickflow/gateway/utils/http"
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
			httpUtils.WriteJSONError(w, errors2.New("UNSUPPORTED_MEDIA_TYE", "Unsupported Content-Type", http.StatusUnsupportedMediaType))
		})
	}
}
