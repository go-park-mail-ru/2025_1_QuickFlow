package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	community_errors "quickflow/community_service/internal/errors"
	"quickflow/shared/models"

	"quickflow/community_service/internal/usecase/mocks"
)

func TestCommunityUseCase_CreateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		community     models.Community
		mockSetup     func(*mocks.MockCommunityRepository, *mocks.MockFileService, *mocks.MockCommunityValidator)
		expected      *models.Community
		expectedError error
	}{
		{
			name: "Success",
			community: models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Test Community",
				},
			},
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Test Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().CreateCommunity(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: &models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Test Community",
				},
			},
		},
		{
			name: "Validation Error",
			community: models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name: "",
				},
			},
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).
					Return(errors.New("validation error"))
			},
			expectedError: errors.New("validation error"),
		},
		{
			name: "Already Exists",
			community: models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Existing Community",
				},
			},
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Existing Community").
					Return(models.Community{ID: uuid.New()}, nil)
			},
			expectedError: community_errors.ErrAlreadyExists,
		},
		{
			name: "With Avatar Upload",
			community: models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Community with Avatar",
				},
				Avatar: &models.File{Name: "avatar.jpg"},
			},
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Community with Avatar").
					Return(models.Community{}, community_errors.ErrNotFound)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return("http://example.com/avatar.jpg", nil)
				repo.EXPECT().CreateCommunity(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: &models.Community{
				BasicInfo: &models.BasicCommunityInfo{
					Name:      "Community with Avatar",
					AvatarUrl: "http://example.com/avatar.jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)
			fs := mocks.NewMockFileService(ctrl)
			val := mocks.NewMockCommunityValidator(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(repo, fs, val)
			}

			uc := NewCommunityUseCase(repo, fs, val)
			result, err := uc.CreateCommunity(context.Background(), tt.community)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.community.BasicInfo.Name, result.BasicInfo.Name)
				if tt.community.Avatar != nil {
					assert.NotEmpty(t, result.BasicInfo.AvatarUrl)
				}
			}
		})
	}
}

func TestCommunityUseCase_GetCommunityById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*mocks.MockCommunityRepository)
		expected      models.Community
		expectedError error
	}{
		{
			name: "Success",
			id:   testID,
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{ID: testID}, nil)
			},
			expected: models.Community{ID: testID},
		},
		{
			name: "Not Found",
			id:   testID,
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{}, community_errors.ErrNotFound)
			},
			expectedError: community_errors.ErrNotFound,
		},
		{
			name:          "Empty ID",
			id:            uuid.Nil,
			mockSetup:     nil,
			expectedError: errors.New("community ID cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			uc := NewCommunityUseCase(repo, nil, nil)
			result, err := uc.GetCommunityById(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
			}
		})
	}
}

func TestCommunityUseCase_UpdateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()
	userID := uuid.New()

	adminRole := models.CommunityRoleAdmin
	memberRole := models.CommunityRoleMember
	ownerRole := models.CommunityRoleOwner

	tests := []struct {
		name          string
		community     models.Community
		userId        uuid.UUID
		mockSetup     func(*mocks.MockCommunityRepository, *mocks.MockFileService, *mocks.MockCommunityValidator)
		expectedError error
	}{
		{
			name: "Success with text update",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Updated Community",
					Description: "New bio",
				},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Updated Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().IsCommunityMember(gomock.Any(), userID, testID).
					Return(true, &adminRole, nil)
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							Name:        "Old Community",
							Description: "Old bio",
						},
					}, nil)
				repo.EXPECT().UpdateCommunityTextInfo(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							Name:        "Updated Community",
							Description: "New bio",
						},
					}, nil)
			},
		},
		{
			name: "Success with avatar update",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Updated Community",
				},
				Avatar: &models.File{Name: "new-avatar.jpg"},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Updated Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().IsCommunityMember(gomock.Any(), userID, testID).
					Return(true, &ownerRole, nil)
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							AvatarUrl: "old-avatar.jpg",
						},
					}, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("http://example.com/new-avatar.jpg", nil)
				repo.EXPECT().UpdateCommunityTextInfo(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().UpdateCommunityAvatar(gomock.Any(), testID, "http://example.com/new-avatar.jpg").
					Return(nil)
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							AvatarUrl: "http://example.com/new-avatar.jpg",
						},
					}, nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "old-avatar.jpg").Return(nil)
			},
		},
		{
			name: "Validation error",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "",
				},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).
					Return(errors.New("name cannot be empty"))
			},
			expectedError: errors.New("name cannot be empty"),
		},
		{
			name: "Community not found",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Updated Community",
				},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Updated Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().IsCommunityMember(gomock.Any(), userID, testID).
					Return(true, &adminRole, nil)
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{}, community_errors.ErrNotFound)
			},
			expectedError: community_errors.ErrNotFound,
		},
		{
			name: "Not a member",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Updated Community",
				},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Updated Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().IsCommunityMember(gomock.Any(), userID, testID).
					Return(false, nil, nil)
			},
			expectedError: community_errors.ErrNotParticipant,
		},
		{
			name: "No permissions (member role)",
			community: models.Community{
				ID: testID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Updated Community",
				},
			},
			userId: userID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, val *mocks.MockCommunityValidator) {
				val.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
				repo.EXPECT().GetCommunityByName(gomock.Any(), "Updated Community").
					Return(models.Community{}, community_errors.ErrNotFound)
				repo.EXPECT().IsCommunityMember(gomock.Any(), userID, testID).
					Return(true, &memberRole, nil)
			},
			expectedError: community_errors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)
			fs := mocks.NewMockFileService(ctrl)
			val := mocks.NewMockCommunityValidator(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(repo, fs, val)
			}

			uc := NewCommunityUseCase(repo, fs, val)
			_, err := uc.UpdateCommunity(context.Background(), tt.community, tt.userId)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommunityUseCase_DeleteCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*mocks.MockCommunityRepository, *mocks.MockFileService)
		expectedError error
	}{
		{
			name: "Success",
			id:   testID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							AvatarUrl: "avatar.jpg",
							CoverUrl:  "cover.jpg",
						},
					}, nil)
				repo.EXPECT().DeleteCommunity(gomock.Any(), testID).Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "avatar.jpg").Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "cover.jpg").Return(nil)
			},
		},
		{
			name: "Community Not Found",
			id:   testID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{}, community_errors.ErrNotFound)
			},
			expectedError: community_errors.ErrNotFound,
		},
		{
			name: "Delete File Error",
			id:   testID,
			mockSetup: func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID: testID,
						BasicInfo: &models.BasicCommunityInfo{
							AvatarUrl: "avatar.jpg",
						},
					}, nil)
				repo.EXPECT().DeleteCommunity(gomock.Any(), testID).Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "avatar.jpg").
					Return(errors.New("delete error"))
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)
			fs := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(repo, fs)
			}

			uc := NewCommunityUseCase(repo, fs, nil)
			err := uc.DeleteCommunity(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommunityUseCase_JoinCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		member        models.CommunityMember
		mockSetup     func(*mocks.MockCommunityRepository)
		expectedError error
	}{
		{
			name: "Success",
			member: models.CommunityMember{
				CommunityID: testID,
				UserID:      userID,
			},
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{ID: testID}, nil)
				repo.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "Community Not Found",
			member: models.CommunityMember{
				CommunityID: testID,
				UserID:      userID,
			},
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{}, community_errors.ErrNotFound)
			},
			expectedError: community_errors.ErrNotFound,
		},
		{
			name: "Already Member",
			member: models.CommunityMember{
				CommunityID: testID,
				UserID:      userID,
			},
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						ID:      testID,
						OwnerID: userID,
					}, nil)
				repo.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).
					Return(community_errors.ErrAlreadyExists)
			},
			expectedError: community_errors.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			uc := NewCommunityUseCase(repo, nil, nil)
			err := uc.JoinCommunity(context.Background(), tt.member)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommunityUseCase_ChangeUserRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()
	userID := uuid.New()
	requesterID := uuid.New()

	admin := models.CommunityRoleAdmin
	member := models.CommunityRoleMember

	tests := []struct {
		name          string
		userId        uuid.UUID
		communityId   uuid.UUID
		role          models.CommunityRole
		requester     uuid.UUID
		mockSetup     func(*mocks.MockCommunityRepository)
		expectedError error
	}{
		{
			name:        "Success Admin",
			userId:      userID,
			communityId: testID,
			role:        models.CommunityRoleAdmin,
			requester:   requesterID,
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().IsCommunityMember(gomock.Any(), requesterID, testID).
					Return(true, &admin, nil)
				repo.EXPECT().ChangeUserRole(gomock.Any(), userID, testID, models.CommunityRoleAdmin).
					Return(nil)
			},
		},
		{
			name:        "Requester Not Admin",
			userId:      userID,
			communityId: testID,
			role:        models.CommunityRoleAdmin,
			requester:   requesterID,
			mockSetup: func(repo *mocks.MockCommunityRepository) {
				repo.EXPECT().IsCommunityMember(gomock.Any(), requesterID, testID).
					Return(true, &member, nil)
			},
			expectedError: community_errors.ErrForbidden,
		},
		{
			name:          "Invalid UUID",
			userId:        uuid.Nil,
			communityId:   testID,
			role:          models.CommunityRoleAdmin,
			requester:     requesterID,
			mockSetup:     nil,
			expectedError: errors.New("user ID and community ID cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockCommunityRepository(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			uc := NewCommunityUseCase(repo, nil, nil)
			err := uc.ChangeUserRole(context.Background(), tt.userId, tt.communityId, tt.role, tt.requester)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
