package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	usererrors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/usecase"
	"quickflow/user_service/internal/usecase/mocks"
)

func TestProfileService_GetUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileRepo := mocks.NewMockFileService(ctrl)

	profileService := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)

	// Используем UUID для теста
	userID := uuid.New()

	tests := []struct {
		name       string
		userId     uuid.UUID
		mockReturn models.Profile
		mockError  error
		expected   models.Profile
		expectErr  bool
	}{
		{
			name:       "Success - Get user info",
			userId:     userID,
			mockReturn: models.Profile{UserId: userID},
			mockError:  nil,
			expected:   models.Profile{UserId: userID},
			expectErr:  false,
		},
		{
			name:       "Error - GetProfile fails",
			userId:     userID,
			mockReturn: models.Profile{},
			mockError:  errors.New("profile not found"),
			expected:   models.Profile{},
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProfileRepo.EXPECT().GetProfile(context.Background(), tt.userId).Return(tt.mockReturn, tt.mockError)

			profile, err := profileService.GetUserInfo(context.Background(), tt.userId)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, profile)
			}
		})
	}
}

func TestProfileService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileRepo := mocks.NewMockFileService(ctrl)

	profileService := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)

	userID := uuid.New()

	tests := []struct {
		name          string
		newProfile    models.Profile
		mockUserError error
		mockUpdateErr error
		expectedErr   bool
		expected      *models.Profile
	}{
		{
			name: "Success - Profile updated",
			newProfile: models.Profile{
				UserId:   userID,
				Username: "newUsername",
			},
			mockUserError: nil,
			mockUpdateErr: nil,
			expectedErr:   false,
			expected:      &models.Profile{UserId: userID, Username: "newUsername"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ожидаем вызов GetUserByUsername
			mockUserRepo.EXPECT().GetUserByUsername(context.Background(), tt.newProfile.Username).Return(models.User{}, tt.mockUserError)

			// Ожидаем вызов UpdateProfileTextInfo с любым контекстом и правильным профилем
			// Добавим проверку контекста в мок
			mockProfileRepo.EXPECT().UpdateProfileTextInfo(gomock.Any(), tt.newProfile).Return(tt.mockUpdateErr)

			// Ожидаем вызов GetProfile с правильным контекстом и UserId
			mockProfileRepo.EXPECT().GetProfile(context.Background(), tt.newProfile.UserId).Return(tt.newProfile, nil)

			// Запуск метода
			profile, err := profileService.UpdateProfile(context.Background(), tt.newProfile)

			// Проверка на ошибку
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, profile)
			}
		})
	}
}

func TestProfileService_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileRepo := mocks.NewMockFileService(ctrl)

	profileService := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)

	tests := []struct {
		name        string
		profile     models.Profile
		mockError   error
		expectedErr bool
		expected    models.Profile
	}{
		{
			name: "Success - Create profile",
			profile: models.Profile{
				UserId:   uuid.New(),
				Username: "newProfile",
			},
			mockError:   nil,
			expectedErr: false,
			expected:    models.Profile{Username: "newProfile"},
		},
		{
			name: "Error - Invalid profile info",
			profile: models.Profile{
				UserId:   uuid.New(),
				Username: "",
			},
			mockError:   usererrors.ErrInvalidProfileInfo,
			expectedErr: true,
			expected:    models.Profile{},
		},
		{
			name: "Error - SaveProfile fails",
			profile: models.Profile{
				UserId:   uuid.New(),
				Username: "newProfile",
			},
			mockError:   errors.New("unable to save profile"),
			expectedErr: true,
			expected:    models.Profile{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProfileRepo.EXPECT().SaveProfile(context.Background(), tt.profile).Return(tt.mockError)

			profile, err := profileService.CreateProfile(context.Background(), tt.profile)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Username, profile.Username)
				assert.NotEqual(t, uuid.Nil, profile.UserId)
			}
		})
	}
}
