package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"

	"quickflow/community_service/internal/delivery/grpc/mocks"
)

func TestCommunityServiceServer_CreateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.CreateCommunityRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.CreateCommunityResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.CreateCommunityRequest{
				Name:        "Test Community",
				Description: "Test Description",
				OwnerId:     uuid.New().String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().CreateCommunity(gomock.Any(), gomock.Any()).
					Return(&models.Community{
						BasicInfo: &models.BasicCommunityInfo{
							Name:        "Test Community",
							Description: "Test Description",
						},
					}, nil)
			},
			expectedResp: &pb.CreateCommunityResponse{
				Community: &pb.Community{
					Name:        "Test Community",
					Description: "Test Description",
				},
			},
		},
		{
			name: "Invalid Owner ID",
			req: &pb.CreateCommunityRequest{
				OwnerId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid owner ID"),
		},
		{
			name: "Create Error",
			req: &pb.CreateCommunityRequest{
				Name:    "Test Community",
				OwnerId: uuid.New().String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().CreateCommunity(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("create error"))
			},
			expectedError: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.CreateCommunity(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Community.Name, resp.Community.Name)
				assert.Equal(t, tt.expectedResp.Community.Description, resp.Community.Description)
			}
		})
	}
}

func TestCommunityServiceServer_GetCommunityById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()

	tests := []struct {
		name          string
		req           *pb.GetCommunityByIdRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.GetCommunityByIdResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.GetCommunityByIdRequest{
				CommunityId: testID.String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{
						BasicInfo: &models.BasicCommunityInfo{
							Name: "Test Community",
						},
					}, nil)
			},
			expectedResp: &pb.GetCommunityByIdResponse{
				Community: &pb.Community{
					Name: "Test Community",
				},
			},
		},
		{
			name: "Invalid ID",
			req: &pb.GetCommunityByIdRequest{
				CommunityId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid community ID"),
		},
		{
			name: "Not Found",
			req: &pb.GetCommunityByIdRequest{
				CommunityId: testID.String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().GetCommunityById(gomock.Any(), testID).
					Return(models.Community{}, errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.GetCommunityById(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Community.Name, resp.Community.Name)
			}
		})
	}
}

func TestCommunityServiceServer_UpdateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		req           *pb.UpdateCommunityRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.UpdateCommunityResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.UpdateCommunityRequest{
				Id:       testID.String(),
				UserId:   userID.String(),
				Name:     "Updated Name",
				Nickname: "New Nickname",
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().UpdateCommunity(gomock.Any(), gomock.Any(), userID).
					Return(&models.Community{
						ID:       testID,
						NickName: "New Nickname",
						BasicInfo: &models.BasicCommunityInfo{
							Name: "Updated Name",
						},
					}, nil)
			},
			expectedResp: &pb.UpdateCommunityResponse{
				Community: &pb.Community{
					Id:       testID.String(),
					Name:     "Updated Name",
					Nickname: "New Nickname",
				},
			},
		},
		{
			name: "Invalid User ID",
			req: &pb.UpdateCommunityRequest{
				UserId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid user ID"),
		},
		{
			name: "Invalid Community ID",
			req: &pb.UpdateCommunityRequest{
				Id:     "invalid-uuid",
				UserId: userID.String(),
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid community ID"),
		},
		{
			name: "Update Error",
			req: &pb.UpdateCommunityRequest{
				Id:     testID.String(),
				UserId: userID.String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().UpdateCommunity(gomock.Any(), gomock.Any(), userID).
					Return(nil, errors.New("update error"))
			},
			expectedError: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.UpdateCommunity(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Community.Name, resp.Community.Name)
				assert.Equal(t, tt.expectedResp.Community.Nickname, resp.Community.Nickname)
			}
		})
	}
}

func TestCommunityServiceServer_DeleteCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()

	tests := []struct {
		name          string
		req           *pb.DeleteCommunityRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.DeleteCommunityResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.DeleteCommunityRequest{
				CommunityId: testID.String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().DeleteCommunity(gomock.Any(), testID).
					Return(nil)
			},
			expectedResp: &pb.DeleteCommunityResponse{},
		},
		{
			name: "Invalid ID",
			req: &pb.DeleteCommunityRequest{
				CommunityId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid community ID"),
		},
		{
			name: "Delete Error",
			req: &pb.DeleteCommunityRequest{
				CommunityId: testID.String(),
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().DeleteCommunity(gomock.Any(), testID).
					Return(errors.New("delete error"))
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.DeleteCommunity(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestCommunityServiceServer_JoinCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		req           *pb.JoinCommunityRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.JoinCommunityResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.JoinCommunityRequest{
				NewMember: &pb.CommunityMember{
					CommunityId: testID.String(),
					UserId:      userID.String(),
					Role:        pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
				},
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedResp: &pb.JoinCommunityResponse{
				Success: true,
			},
		},
		{
			name: "Invalid Member Data",
			req: &pb.JoinCommunityRequest{
				NewMember: &pb.CommunityMember{
					CommunityId: "invalid-uuid",
				},
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid member data"),
		},
		{
			name: "Join Error",
			req: &pb.JoinCommunityRequest{
				NewMember: &pb.CommunityMember{
					CommunityId: testID.String(),
					UserId:      userID.String(),
				},
			},
			mockSetup: func(uc *mocks.MockCommunityUseCase) {
				uc.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).
					Return(errors.New("join error"))
			},
			expectedError: errors.New("join error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.JoinCommunity(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestCommunityServiceServer_GetCommunityMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := uuid.New()

	tests := []struct {
		name          string
		req           *pb.GetCommunityMembersRequest
		mockSetup     func(*mocks.MockCommunityUseCase)
		expectedResp  *pb.GetCommunityMembersResponse
		expectedError error
	}{
		{
			name: "Invalid Community ID",
			req: &pb.GetCommunityMembersRequest{
				CommunityId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid community ID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockCommunityUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := NewCommunityServiceServer(mockUC)
			resp, err := server.GetCommunityMembers(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Members, 1)
				assert.Equal(t, testID.String(), resp.Members[0].CommunityId)
			}
		})
	}
}
