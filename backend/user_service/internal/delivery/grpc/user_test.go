package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	dto "quickflow/shared/client/user_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"

	"quickflow/user_service/internal/delivery/grpc/mocks"
)

func TestUserServiceServer_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.SignUpRequest
		mockSetup     func(*mocks.MockUserUseCase, *pb.SignUpRequest)
		expectedResp  *pb.SignUpResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Id:       uuid.New().String(),
					Username: "testuser",
					Password: "testpass",
				},
				Profile: &pb.Profile{
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "Test",
						Lastname:  "User",
					},
				},
			},
			mockSetup: func(m *mocks.MockUserUseCase, req *pb.SignUpRequest) {
				user, _ := dto.MapUserDTOToUser(req.User)
				profile, _ := dto.MapProfileDTOToProfile(req.Profile)
				session := models.Session{SessionId: uuid.New()}
				m.EXPECT().CreateUser(gomock.Any(), *user, *profile).
					Return(uuid.New(), session, nil)
			},
			expectedResp: &pb.SignUpResponse{
				Session: &pb.Session{Id: gomock.Any().String()},
			},
		},
		{
			name: "Invalid User Data",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Username: "testuser",
					Password: "testpass",
				},
				Profile: &pb.Profile{
					Username: "testuser",
				},
			},
			mockSetup:     nil,
			expectedError: status.Error(codes.InvalidArgument, "invalid user data"),
		},
		{
			name: "Create User Error",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Id:       uuid.New().String(),
					Username: "testuser",
					Password: "testpass",
				},
				Profile: &pb.Profile{
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "Test",
						Lastname:  "User",
					},
				},
			},
			mockSetup: func(m *mocks.MockUserUseCase, req *pb.SignUpRequest) {
				user, _ := dto.MapUserDTOToUser(req.User)
				profile, _ := dto.MapProfileDTOToProfile(req.Profile)
				m.EXPECT().CreateUser(gomock.Any(), *user, *profile).
					Return(uuid.Nil, models.Session{}, errors.New("create error"))
			},
			expectedError: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUseCase, tt.req)
			}

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.SignUp(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if statusErr, ok := status.FromError(err); ok {
					assert.Contains(t, statusErr.Message(), "invalid user data")
				} else {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				if resp != nil {
					assert.NotNil(t, resp)
					assert.NotEmpty(t, resp.Session.Id)
				}
			}
		})
	}
}

func TestUserServiceServer_SignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.SignInRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.SignInResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.SignInRequest{
				SignIn: &pb.SignIn{
					Username: "testuser",
					Password: "testpass",
				},
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				session := models.Session{SessionId: uuid.New()}
				m.EXPECT().AuthUser(gomock.Any(), models.LoginData{
					Username: "testuser",
					Password: "testpass",
				}).Return(session, nil)
			},
			expectedResp: &pb.SignInResponse{
				Session: &pb.Session{Id: gomock.Any().String()},
			},
		},
		{
			name: "Auth Error",
			req: &pb.SignInRequest{
				SignIn: &pb.SignIn{
					Username: "testuser",
					Password: "wrongpass",
				},
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().AuthUser(gomock.Any(), gomock.Any()).
					Return(models.Session{}, errors.New("auth error"))
			},
			expectedError: errors.New("auth error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.SignIn(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Session.Id)
			}
		})
	}
}

func TestUserServiceServer_SignOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.SignOutRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.SignOutResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.SignOutRequest{
				SessionId: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().DeleteUserSession(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedResp: &pb.SignOutResponse{Success: true},
		},
		{
			name: "Delete Error",
			req: &pb.SignOutRequest{
				SessionId: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().DeleteUserSession(gomock.Any(), gomock.Any()).
					Return(errors.New("delete error"))
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.SignOut(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.False(t, resp.Success)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestUserServiceServer_GetUserByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.GetUserByUsernameRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.GetUserByUsernameResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.GetUserByUsernameRequest{
				Username: "testuser",
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				user := models.User{
					Id:       uuid.New(),
					Username: "testuser",
				}
				m.EXPECT().GetUserByUsername(gomock.Any(), "testuser").
					Return(user, nil)
			},
			expectedResp: &pb.GetUserByUsernameResponse{
				User: &pb.User{
					Id:       gomock.Any().String(),
					Username: "testuser",
				},
			},
		},
		{
			name: "User Not Found",
			req: &pb.GetUserByUsernameRequest{
				Username: "nonexistent",
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().GetUserByUsername(gomock.Any(), "nonexistent").
					Return(models.User{}, errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.GetUserByUsername(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.req.Username, resp.User.Username)
				assert.NotEmpty(t, resp.User.Id)
			}
		})
	}
}

func TestUserServiceServer_GetUserById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.GetUserByIdRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.GetUserByIdResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.GetUserByIdRequest{
				Id: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				userId := uuid.New()
				user := models.User{
					Id:       userId,
					Username: "testuser",
				}
				m.EXPECT().GetUserById(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, id uuid.UUID) (models.User, error) {
						assert.Equal(t, userId, id)
						return user, nil
					})
			},
			expectedResp: &pb.GetUserByIdResponse{
				User: &pb.User{
					Id:       gomock.Any().String(),
					Username: "testuser",
				},
			},
		},
		{
			name: "Invalid UUID",
			req: &pb.GetUserByIdRequest{
				Id: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid user id"),
		},
		{
			name: "User Not Found",
			req: &pb.GetUserByIdRequest{
				Id: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().GetUserById(gomock.Any(), gomock.Any()).
					Return(models.User{}, errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUseCase)
			}

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.GetUserById(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.User.Id)
				assert.Equal(t, "testuser", resp.User.Username)
			}
		})
	}
}

func TestUserServiceServer_LookupUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		req           *pb.LookupUserSessionRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.LookupUserSessionResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.LookupUserSessionRequest{
				SessionId: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				sessionId := uuid.New()
				user := models.User{
					Id:       uuid.New(),
					Username: "testuser",
				}
				m.EXPECT().LookupUserSession(gomock.Any(), models.Session{SessionId: sessionId}).
					Return(user, nil)
			},
			expectedResp: &pb.LookupUserSessionResponse{
				UserId:   gomock.Any().String(),
				Username: "testuser",
			},
		},
		{
			name: "Invalid Session ID",
			req: &pb.LookupUserSessionRequest{
				SessionId: "invalid-uuid",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid session id"),
		},
		{
			name: "Session Not Found",
			req: &pb.LookupUserSessionRequest{
				SessionId: uuid.New().String(),
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().LookupUserSession(gomock.Any(), gomock.Any()).
					Return(models.User{}, errors.New("session not found"))
			},
			expectedError: errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUseCase)
			}

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.LookupUserSession(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.UserId)
				assert.Equal(t, "testuser", resp.Username)
			}
		})
	}
}

func TestUserServiceServer_SearchSimilarUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	protoNow := timestamppb.New(now)

	tests := []struct {
		name          string
		req           *pb.SearchSimilarUserRequest
		mockSetup     func(*mocks.MockUserUseCase)
		expectedResp  *pb.SearchSimilarUserResponse
		expectedError error
	}{
		{
			name: "Success",
			req: &pb.SearchSimilarUserRequest{
				ToSearch: "test",
				NumUsers: 5,
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				users := []models.PublicUserInfo{
					{
						Id:        uuid.New(),
						Username:  "testuser1",
						Firstname: "Test1",
						Lastname:  "User1",
						LastSeen:  now,
					},
					{
						Id:        uuid.New(),
						Username:  "testuser2",
						Firstname: "Test2",
						Lastname:  "User2",
						LastSeen:  now,
					},
				}
				m.EXPECT().SearchSimilarUser(gomock.Any(), "test", uint(5)).
					Return(users, nil)
			},
			expectedResp: &pb.SearchSimilarUserResponse{
				UsersInfo: []*pb.PublicUserInfo{
					{
						Id:        gomock.Any().String(),
						Username:  "testuser1",
						Firstname: "Test1",
						Lastname:  "User1",
						LastSeen:  protoNow,
					},
					{
						Id:        gomock.Any().String(),
						Username:  "testuser2",
						Firstname: "Test2",
						Lastname:  "User2",
						LastSeen:  protoNow,
					},
				},
			},
		},
		{
			name: "Search Error",
			req: &pb.SearchSimilarUserRequest{
				ToSearch: "test",
				NumUsers: 5,
			},
			mockSetup: func(m *mocks.MockUserUseCase) {
				m.EXPECT().SearchSimilarUser(gomock.Any(), "test", uint(5)).
					Return(nil, errors.New("search error"))
			},
			expectedError: errors.New("search error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockUserUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewUserServiceServer(mockUseCase)
			resp, err := server.SearchSimilarUser(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.UsersInfo, 2)
				assert.Equal(t, "testuser1", resp.UsersInfo[0].Username)
				assert.Equal(t, "Test1", resp.UsersInfo[0].Firstname)
				assert.Equal(t, "User1", resp.UsersInfo[0].Lastname)
				assert.Equal(t, "testuser2", resp.UsersInfo[1].Username)
				assert.Equal(t, "Test2", resp.UsersInfo[1].Firstname)
				assert.Equal(t, "User2", resp.UsersInfo[1].Lastname)
				assert.NotNil(t, resp.UsersInfo[0].LastSeen)
				assert.NotNil(t, resp.UsersInfo[1].LastSeen)
			}
		})
	}
}
