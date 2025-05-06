package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/friends_service/internal/usecase"
	"quickflow/friends_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestFriendsService(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFriendsRepository(ctrl)
	service := usecase.NewFriendsService(mockRepo)

	ctx := context.Background()
	validUserID := uuid.New().String()
	validFriendID := uuid.New().String()

	t.Run("GetFriendsInfo", func(t *testing.T) {
		tests := []struct {
			name          string
			userID        string
			limit         string
			offset        string
			reqType       string
			mockSetup     func()
			expectedResp  []models.FriendInfo
			expectedCount int
			expectedErr   error
		}{
			{
				name:    "Success",
				userID:  validUserID,
				limit:   "10",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					mockRepo.EXPECT().GetFriendsPublicInfo(ctx, validUserID, 10, 0, "all").
						Return([]models.FriendInfo{
							{Id: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Username: "user1"},
						}, 1, nil)
				},
				expectedResp: []models.FriendInfo{
					{Id: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Username: "user1"},
				},
				expectedCount: 1,
				expectedErr:   nil,
			},
			// In the GetFriendsInfo test cases, update the expected errors:
			{
				name:    "Invalid limit",
				userID:  validUserID,
				limit:   "invalid",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					// No mock expectations needed as it fails before repo call
				},
				expectedResp:  nil,
				expectedCount: 0,
				expectedErr:   fmt.Errorf("strconv.Atoi: parsing \"invalid\": invalid syntax"),
			},
			{
				name:    "Repository error",
				userID:  validUserID,
				limit:   "10",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					mockRepo.EXPECT().GetFriendsPublicInfo(ctx, validUserID, 10, 0, "all").
						Return(nil, 0, errors.New("db error"))
				},
				expectedResp:  []models.FriendInfo{},
				expectedCount: 0,
				expectedErr:   errors.New("db error"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}

				resp, count, err := service.GetFriendsInfo(ctx, tt.userID, tt.limit, tt.offset, tt.reqType)

				if tt.expectedErr != nil {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tt.expectedResp, resp)
				assert.Equal(t, tt.expectedCount, count)
			})
		}
	})

	t.Run("SendFriendRequest", func(t *testing.T) {
		tests := []struct {
			name        string
			senderID    string
			receiverID  string
			mockSetup   func()
			expectedErr error
		}{
			{
				name:       "Success",
				senderID:   validUserID,
				receiverID: validFriendID,
				mockSetup: func() {
					mockRepo.EXPECT().SendFriendRequest(ctx, validUserID, validFriendID).
						Return(nil)
				},
				expectedErr: nil,
			},
			{
				name:       "Repository error",
				senderID:   validUserID,
				receiverID: validFriendID,
				mockSetup: func() {
					mockRepo.EXPECT().SendFriendRequest(ctx, validUserID, validFriendID).
						Return(errors.New("db error"))
				},
				expectedErr: errors.New("db error"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}

				err := service.SendFriendRequest(ctx, tt.senderID, tt.receiverID)

				if tt.expectedErr != nil {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("GetUserRelation", func(t *testing.T) {
		user1 := uuid.New()
		user2 := uuid.New()

		tests := []struct {
			name        string
			user1       uuid.UUID
			user2       uuid.UUID
			mockSetup   func()
			expectedRel models.UserRelation
			expectedErr error
		}{
			{
				name:  "Success - Friends",
				user1: user1,
				user2: user2,
				mockSetup: func() {
					mockRepo.EXPECT().GetUserRelation(ctx, user1, user2).
						Return(models.RelationFriend, nil)
				},
				expectedRel: models.RelationFriend,
				expectedErr: nil,
			},
			{
				name:  "Same user",
				user1: user1,
				user2: user1,
				mockSetup: func() {
					// No repo call expected
				},
				expectedRel: models.RelationSelf,
				expectedErr: nil,
			},
			{
				name:  "Nil UUID",
				user1: uuid.Nil,
				user2: user2,
				mockSetup: func() {
					// No repo call expected
				},
				expectedRel: models.RelationStranger,
				expectedErr: fmt.Errorf("userID is empty"),
			},
			{
				name:  "Repository error",
				user1: user1,
				user2: user2,
				mockSetup: func() {
					mockRepo.EXPECT().GetUserRelation(ctx, user1, user2).
						Return(models.RelationStranger, errors.New("db error"))
				},
				expectedRel: models.RelationStranger,
				expectedErr: fmt.Errorf("f.friendsRepo.GetUserRelation"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}

				rel, err := service.GetUserRelation(ctx, tt.user1, tt.user2)

				if tt.expectedErr != nil {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tt.expectedRel, rel)
			})
		}
	})

	t.Run("AcceptFriendRequest", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo.EXPECT().AcceptFriendRequest(ctx, validUserID, validFriendID).
				Return(nil)

			err := service.AcceptFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(t, err)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo.EXPECT().AcceptFriendRequest(ctx, validUserID, validFriendID).
				Return(errors.New("db error"))

			err := service.AcceptFriendRequest(ctx, validUserID, validFriendID)
			assert.Error(t, err)
		})
	})

	t.Run("DeleteFriend", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo.EXPECT().DeleteFriend(ctx, validUserID, validFriendID).
				Return(nil)

			err := service.DeleteFriend(ctx, validUserID, validFriendID)
			assert.NoError(t, err)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo.EXPECT().DeleteFriend(ctx, validUserID, validFriendID).
				Return(errors.New("db error"))

			err := service.DeleteFriend(ctx, validUserID, validFriendID)
			assert.Error(t, err)
		})
	})

	t.Run("Unfollow", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo.EXPECT().Unfollow(ctx, validUserID, validFriendID).
				Return(nil)

			err := service.Unfollow(ctx, validUserID, validFriendID)
			assert.NoError(t, err)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo.EXPECT().Unfollow(ctx, validUserID, validFriendID).
				Return(errors.New("db error"))

			err := service.Unfollow(ctx, validUserID, validFriendID)
			assert.Error(t, err)
		})
	})

	t.Run("IsExistsFriendRequest", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).
				Return(true, nil)

			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(t, err)
			assert.True(t, exists)
		})

		t.Run("Not exists", func(t *testing.T) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).
				Return(false, nil)

			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(t, err)
			assert.False(t, exists)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).
				Return(false, errors.New("db error"))

			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.Error(t, err)
			assert.False(t, exists)
		})
	})
}
