package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	forms2 "quickflow/monolith/internal/delivery/forms"
	http2 "quickflow/monolith/internal/delivery/http"
	"quickflow/monolith/internal/delivery/http/mocks"
	"quickflow/monolith/internal/models"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchHandler_SearchSimilar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearchUseCase := mocks.NewMockSearchUseCase(ctrl)
	handler := http2.NewSearchHandler(mockSearchUseCase)

	// Test users
	testUser1 := models.PublicUserInfo{
		Id:        uuid.New(),
		Username:  "testuser1",
		Firstname: "Test",
		Lastname:  "User1",
		AvatarURL: "http://example.com/avatar1.jpg",
	}
	testUser2 := models.PublicUserInfo{
		Id:        uuid.New(),
		Username:  "testuser2",
		Firstname: "Test",
		Lastname:  "User2",
	}

	tests := []struct {
		name           string
		queryParams    url.Values
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Successful search with results",
			queryParams: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"5"},
			},
			mockSetup: func() {
				mockSearchUseCase.EXPECT().
					SearchSimilarUser(gomock.Any(), "test", uint(5)).
					Return([]models.PublicUserInfo{testUser1, testUser2}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response forms2.PayloadWrapper[[]forms2.PublicUserInfoOut]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				require.Len(t, response.Payload, 2)
				assert.Equal(t, testUser1.Id.String(), response.Payload[0].ID)
				assert.Equal(t, testUser2.Id.String(), response.Payload[1].ID)
			},
		},
		{
			name: "Missing required parameter 'string'",
			queryParams: url.Values{
				"users_count": []string{"5"},
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Failed to decode request body",
		},
		{
			name: "Missing required parameter 'users_count'",
			queryParams: url.Values{
				"string": []string{"test"},
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Failed to decode request body",
		},
		{
			name: "Invalid users_count format",
			queryParams: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"invalid"},
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Failed to decode request body",
		},
		{
			name: "Search returns error",
			queryParams: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"5"},
			},
			mockSetup: func() {
				mockSearchUseCase.EXPECT().
					SearchSimilarUser(gomock.Any(), "test", uint(5)).
					Return(nil, errors.New("search error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to search similar users",
		},
		{
			name: "Empty search results",
			queryParams: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"5"},
			},
			mockSetup: func() {
				mockSearchUseCase.EXPECT().
					SearchSimilarUser(gomock.Any(), "test", uint(5)).
					Return([]models.PublicUserInfo{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response forms2.PayloadWrapper[[]forms2.PublicUserInfoOut]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Empty(t, response.Payload)
			},
		},
		{
			name: "Users count exceeds maximum",
			queryParams: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"1000"},
			},
			mockSetup: func() {
				// Assuming the use case handles the maximum limit internally
				mockSearchUseCase.EXPECT().
					SearchSimilarUser(gomock.Any(), "test", uint(1000)).
					Return([]models.PublicUserInfo{testUser1}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response forms2.PayloadWrapper[[]forms2.PublicUserInfoOut]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				require.Len(t, response.Payload, 1)
				assert.Equal(t, testUser1.Id.String(), response.Payload[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.mockSetup()

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/search?"+tt.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			// Call handler
			handler.SearchSimilar(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			if tt.expectedBody != nil {
				assert.Contains(t, w.Body.String(), tt.expectedBody.(string))
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestSearchForm_Unpack(t *testing.T) {
	tests := []struct {
		name        string
		values      url.Values
		expected    forms2.SearchForm
		expectError bool
	}{
		{
			name: "Valid form",
			values: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"10"},
			},
			expected: forms2.SearchForm{
				ToSearch:   "test",
				UsersCount: 10,
			},
			expectError: false,
		},
		{
			name: "Missing string parameter",
			values: url.Values{
				"users_count": []string{"10"},
			},
			expectError: true,
		},
		{
			name: "Missing users_count parameter",
			values: url.Values{
				"string": []string{"test"},
			},
			expectError: true,
		},
		{
			name: "Invalid users_count format",
			values: url.Values{
				"string":      []string{"test"},
				"users_count": []string{"invalid"},
			},
			expectError: true,
		},
		{
			name: "Empty string parameter",
			values: url.Values{
				"string":      []string{""},
				"users_count": []string{"10"},
			},
			expected: forms2.SearchForm{
				ToSearch:   "",
				UsersCount: 10,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form forms2.SearchForm
			err := form.Unpack(tt.values)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, form)
			}
		})
	}
}
