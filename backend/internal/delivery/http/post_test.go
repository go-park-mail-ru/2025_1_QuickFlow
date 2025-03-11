package http

import (
    "bytes"
    "context"
    "errors"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "net/http"
    "net/http/httptest"
    "quickflow/internal/models"
    http2 "quickflow/test/http"
    "testing"
)

func TestFeedHandler_AddPost_TableDriven(t *testing.T) {
    mockPostUseCase := new(http2.MockPostUseCase)
    mockAuthUseCase := new(http2.MockAuthUseCase)

    handler := NewFeedHandler(mockPostUseCase, mockAuthUseCase)

    userID := uuid.New()
    user := models.User{Id: userID}

    tests := []struct {
        name           string
        inputPostForm  string // JSON строка для теста
        contextUser    *models.User
        mockError      error
        expectedStatus int
    }{
        {
            name: "Successful post addition",
            inputPostForm: `{
				"desc": "Hello, world!",
				"pics": ["pic1.jpg", "pic2.jpg"]
			}`,
            contextUser:    &user,
            mockError:      nil,
            expectedStatus: http.StatusOK,
        },
        {
            name: "Failed to add post (internal error)",
            inputPostForm: `{
				"desc": "This should fail",
				"pics": ["pic3.jpg"]
			}`,
            contextUser:    &user,
            mockError:      errors.New("database error"),
            expectedStatus: http.StatusInternalServerError,
        },
        {
            name:           "Invalid JSON request",
            inputPostForm:  `invalid json`,
            contextUser:    &user,
            mockError:      nil,
            expectedStatus: http.StatusBadRequest,
        },
        {
            name: "Missing user in context",
            inputPostForm: `{
				"desc": "Should fail due to missing user"
			}`,
            contextUser:    nil, // Пользователь отсутствует
            mockError:      nil,
            expectedStatus: http.StatusInternalServerError,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockPostUseCase.ExpectedCalls = nil

            req := httptest.NewRequest(http.MethodPost, "/post", bytes.NewBufferString(tt.inputPostForm))

            if tt.contextUser != nil {
                req = req.WithContext(context.WithValue(req.Context(), "user", *tt.contextUser))
            }

            w := httptest.NewRecorder()

            if tt.contextUser != nil {
                mockPostUseCase.On("AddPost", mock.Anything, mock.Anything).
                    Return(tt.mockError).
                    Maybe()
            }

            handler.AddPost(w, req)

            resp := w.Result()
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)

            mockPostUseCase.AssertExpectations(t)
        })
    }
}
