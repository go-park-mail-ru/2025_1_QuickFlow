package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	myhttp "quickflow/gateway/internal/delivery/http"
)

func TestGetCSRF_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/csrf", nil)

	handler := &myhttp.CSRFHandler{}
	handler.GetCSRF(rr, req)

	tokenHeader := rr.Header().Get("X-CSRF-Token")
	assert.NotEmpty(t, tokenHeader, "X-CSRF-Token header should not be empty")

	cookies := rr.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			csrfCookie = cookie
			break
		}
	}
	assert.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.Equal(t, tokenHeader, csrfCookie.Value, "Cookie value should match X-CSRF-Token header")
}
