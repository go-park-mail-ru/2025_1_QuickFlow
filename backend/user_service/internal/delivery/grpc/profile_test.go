package grpc_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"quickflow/shared/models"
	"quickflow/shared/proto/user_service"
	"quickflow/user_service/internal/delivery/grpc"
	"quickflow/user_service/internal/delivery/grpc/mocks"
	"quickflow/user_service/internal/errors"
)

func TestProfileService_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	service := grpc.NewProfileServiceServer(mockProfileUC)

	testCases := []struct {
		name        string
		request     *user_service.CreateProfileRequest
		mockCall    func()
		expectedErr error
	}{
		{
			name: "success",
			request: &user_service.CreateProfileRequest{
				Profile: &user_service.Profile{
					Username: "john_doe",
					BasicInfo: &user_service.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe",
					},
				},
			},
			mockCall: func() {
				// Генерация UUID
				mockProfileUC.EXPECT().CreateProfile(gomock.Any(), gomock.Any()).Return(models.Profile{
					UserId: uuid.New(), // Генерация корректного UUID
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid profile data",
			request: &user_service.CreateProfileRequest{
				Profile: nil,
			},
			mockCall:    func() {},
			expectedErr: errors.ErrInvalidProfileInfo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockCall()

			_, err := service.CreateProfile(context.Background(), tc.request)
			if err != nil && err != tc.expectedErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestProfileService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	service := grpc.NewProfileServiceServer(mockProfileUC)

	testCases := []struct {
		name        string
		request     *user_service.UpdateProfileRequest
		mockCall    func()
		expectedErr error
	}{
		{
			name: "success",
			request: &user_service.UpdateProfileRequest{
				Profile: &user_service.Profile{
					Username: "john_doe_updated",
					BasicInfo: &user_service.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe Updated",
					},
				},
			},
			mockCall: func() {
				// Генерация UUID
				mockProfileUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(&models.Profile{
					UserId: uuid.New(), // Генерация корректного UUID
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid profile data",
			request: &user_service.UpdateProfileRequest{
				Profile: nil,
			},
			mockCall:    func() {},
			expectedErr: errors.ErrInvalidProfileInfo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockCall()

			_, err := service.UpdateProfile(context.Background(), tc.request)
			if err != nil && err != tc.expectedErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestProfileService_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	service := grpc.NewProfileServiceServer(mockProfileUC)

	testCases := []struct {
		name        string
		request     *user_service.GetProfileRequest
		mockCall    func()
		expectedErr error
	}{
		{
			name: "success",
			request: &user_service.GetProfileRequest{
				UserId: uuid.New().String(),
			},
			mockCall: func() {
				mockProfileUC.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(models.Profile{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid user id",
			request: &user_service.GetProfileRequest{
				UserId: "invalid-uuid",
			},
			mockCall:    func() {},
			expectedErr: errors.ErrInvalidUserId,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockCall()

			_, err := service.GetProfile(context.Background(), tc.request)
			if err != nil && err != tc.expectedErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestProfileService_GetProfileByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
	service := grpc.NewProfileServiceServer(mockProfileUC)

	testCases := []struct {
		name        string
		request     *user_service.GetProfileByUsernameRequest
		mockCall    func()
		expectedErr error
	}{
		{
			name: "success",
			request: &user_service.GetProfileByUsernameRequest{
				Username: "john_doe",
			},
			mockCall: func() {
				mockProfileUC.EXPECT().GetProfileByUsername(gomock.Any(), gomock.Any()).Return(models.Profile{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid username",
			request: &user_service.GetProfileByUsernameRequest{
				Username: "",
			},
			mockCall:    func() {},
			expectedErr: errors.ErrUserValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockCall()

			_, err := service.GetProfileByUsername(context.Background(), tc.request)
			if err != nil && err != tc.expectedErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}
