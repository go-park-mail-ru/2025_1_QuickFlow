package http_test

import (
	"net/http"
	"net/http/httptest"
	http2 "quickflow/utils/http"
	"strings"
	"testing"
)

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		statusCode   int
		expectedBody string
		expectedCode int
	}{
		{
			name:         "Bad Request",
			message:      "Bad Request",
			statusCode:   http.StatusBadRequest,
			expectedBody: `{"error":"Bad Request"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Unauthorized",
			message:      "Unauthorized",
			statusCode:   http.StatusUnauthorized,
			expectedBody: `{"error":"Unauthorized"}`,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём новый httptest.ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем WriteJSONError
			http2.WriteJSONError(rr, tt.message, tt.statusCode)

			// Проверяем статус-код
			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("WriteJSONError() status code = %v, want %v", status, tt.expectedCode)
			}

			// Проверяем тело ответа, убирая символы новой строки и пробелы в конце
			if body := strings.TrimSpace(rr.Body.String()); body != tt.expectedBody {
				t.Errorf("WriteJSONError() body = %v, want %v", body, tt.expectedBody)
			}
		})
	}
}
