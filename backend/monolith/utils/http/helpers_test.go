package http_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	customErr "quickflow/monolith/utils/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSONError_Table(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		statusCode int
	}{
		{"bad request", "bad", http.StatusBadRequest},
		{"not found", "not found", http.StatusNotFound},
		{"internal error", "fail", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			customErr.WriteJSONError(rec, tt.msg, tt.statusCode)

			res := rec.Result()
			defer res.Body.Close()

			require.Equal(t, tt.statusCode, res.StatusCode)
			require.Equal(t, "application/json", res.Header.Get("Content-Type"))

			var resp map[string]string
			err := json.NewDecoder(res.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, tt.msg, resp["error"])
		})
	}
}

func createMultipartRequest(t *testing.T, fieldName string, files map[string]string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for filename, content := range files {
		part, err := writer.CreateFormFile(fieldName, filename)
		require.NoError(t, err)
		_, err = part.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(15<<20))

	return req
}

func TestGetFiles_Table(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		files       map[string]string
		wantNames   []string
		expectError bool
	}{
		{
			"two valid files",
			"docs",
			map[string]string{"a.txt": "a", "b.md": "b"},
			[]string{"a.txt", "b.md"},
			false,
		},
		{
			"no files in field",
			"none",
			map[string]string{},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMultipartRequest(t, tt.field, tt.files)
			files, err := customErr.GetFiles(req, tt.field)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if len(tt.wantNames) == 0 {
					require.Len(t, files, 0)
				} else {
					var actualNames []string
					for _, f := range files {
						actualNames = append(actualNames, f.Name)
					}
					require.ElementsMatch(t, tt.wantNames, actualNames)
				}
			}
		})
	}
}
