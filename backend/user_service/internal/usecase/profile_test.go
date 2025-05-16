package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
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
