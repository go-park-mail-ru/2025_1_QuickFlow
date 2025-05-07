package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time2 "quickflow/config/time"
	"quickflow/gateway/internal/delivery/forms"
	http2 "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestFeedHandler_GetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUseCase := mocks.NewMockPostUseCase(ctrl)
	mockAuthUseCase := mocks.NewMockAuthUseCase(ctrl)
	mockFriendsUseCase := mocks.NewMockFriendsUseCase(ctrl)
	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	handler := http2.NewFeedHandler(mockAuthUseCase, mockPostUseCase, mockProfileUC, mockFriendsUseCase)

	user := models.User{Id: uuid.New()}
	now := time.Now()
	nowStr := now.Format(time2.TimeStampLayout)

	tests := []struct {
		name           string
		queryParams    url.Values
		mockSetup      func(*testing.T)
		expectedStatus int
		expectedLen    int
		passUser       bool
	}{
		{
			name: "Successful fetch with 2 posts",
			queryParams: url.Values{
				"posts_count": []string{"2"},
				"ts":          []string{nowStr},
			},
			mockSetup: func(t *testing.T) {
				post1 := models.Post{CreatorId: uuid.New(), Desc: "Post 1"}
				post2 := models.Post{CreatorId: uuid.New(), Desc: "Post 2"}

				mockPostUseCase.EXPECT().
					FetchFeed(gomock.Any(), user, 2, gomock.Any()).
					Return([]models.Post{post1, post2}, nil)

				// Use gomock.Any() for the UUID slice since we can't predict the order
				mockProfileUC.EXPECT().
					GetPublicUsersInfo(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.PublicUserInfo, error) {
						// Verify we got exactly 2 unique UUIDs
						assert.Len(t, ids, 2)
						assert.NotEqual(t, ids[0], ids[1])

						return map[uuid.UUID]models.PublicUserInfo{
							ids[0]: {Id: ids[0]},
							ids[1]: {Id: ids[1]},
						}, nil
					})

				// Expect two relation checks with any UUID
				mockFriendsUseCase.EXPECT().
					GetUserRelation(gomock.Any(), user.Id, gomock.Any()).
					Return(models.RelationNone, nil).
					Times(2)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
			passUser:       true,
		},
		{
			name: "FetchFeed returns error",
			queryParams: url.Values{
				"posts_count": []string{"3"},
				"ts":          []string{nowStr},
			},
			mockSetup: func(t *testing.T) {
				mockPostUseCase.EXPECT().
					FetchFeed(gomock.Any(), user, 3, gomock.Any()).
					Return(nil, errors.New("fetch error"))
				// No author info needed since error occurs first
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
			passUser:       true,
		},
		{
			name: "Invalid request format (missing posts_count)",
			queryParams: url.Values{
				"ts": []string{nowStr},
			},
			mockSetup:      func(t *testing.T) {},
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
			passUser:       true,
		},
		{
			name: "No user in context",
			queryParams: url.Values{
				"posts_count": []string{"2"},
				"ts":          []string{nowStr},
			},
			mockSetup:      func(t *testing.T) {},
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
			passUser:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/feed?"+tt.queryParams.Encode(), nil)
			req.Header.Set("Content-Type", "application/json")

			if tt.passUser {
				req = req.WithContext(context.WithValue(req.Context(), "user", user))
			}

			w := httptest.NewRecorder()

			tt.mockSetup(t)
			handler.GetFeed(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var posts []forms.PostOut
				err := json.NewDecoder(resp.Body).Decode(&posts)
				assert.NoError(t, err)
				assert.Len(t, posts, tt.expectedLen)
			}
		})
	}
}
