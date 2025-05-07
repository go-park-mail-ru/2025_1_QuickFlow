package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dto "quickflow/shared/client/user_service"
	userclient "quickflow/shared/client/user_service"
	"quickflow/shared/logger"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, user shared_models.User, profile shared_models.Profile) (uuid.UUID, shared_models.Session, error)
	AuthUser(ctx context.Context, authData shared_models.LoginData) (shared_models.Session, error)
	GetUserByUsername(ctx context.Context, username string) (shared_models.User, error)
	LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error)
	DeleteUserSession(ctx context.Context, session string) error
	GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error)
	SearchSimilarUser(ctx context.Context, toSearch string, usersCount uint) ([]shared_models.PublicUserInfo, error)
}

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	authUseCase UserUseCase
}

func NewUserServiceServer(authUseCase UserUseCase) *UserServiceServer {
	return &UserServiceServer{authUseCase: authUseCase}
}

func (s *UserServiceServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	logger.Info(ctx, "SignUp called")
	user, err := dto.MapUserDTOToUser(req.User)
	if err != nil {
		logger.Error(ctx, "invalid user data:", err)
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	profile, err := dto.MapProfileDTOToProfile(req.Profile)
	if err != nil {
		logger.Error(ctx, "invalid profile data:", err)
		return nil, fmt.Errorf("invalid profile data: %w", err)
	}

	_, session, err := s.authUseCase.CreateUser(ctx, *user, *profile)
	if err != nil {
		logger.Error(ctx, "failed to create user:", err)
		return nil, err
	}

	return &pb.SignUpResponse{Session: dto.MapSessionToDTO(session)}, nil
}

func (s *UserServiceServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	logger.Info(ctx, "SignIn called")
	loginData := dto.MapSignInToSignInDTO(req.SignIn)

	session, err := s.authUseCase.AuthUser(ctx, *loginData)
	if err != nil {
		logger.Error(ctx, "failed to authenticate user:", err)
		return nil, err
	}

	return &pb.SignInResponse{Session: dto.MapSessionToDTO(session)}, nil
}

func (s *UserServiceServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	logger.Info(ctx, "SignOut called")
	err := s.authUseCase.DeleteUserSession(ctx, req.SessionId)
	if err != nil {
		logger.Error(ctx, "failed to sign out:", err)
		return &pb.SignOutResponse{Success: false}, err
	}
	return &pb.SignOutResponse{Success: true}, nil
}

func (s *UserServiceServer) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserByUsernameResponse, error) {
	logger.Info(ctx, "GetUserByUsername called:", req.Username)
	user, err := s.authUseCase.GetUserByUsername(ctx, req.Username)
	if err != nil {
		logger.Error(ctx, "failed to get user by username:", err)
		return nil, err
	}
	return &pb.GetUserByUsernameResponse{User: dto.MapUserToUserDTO(&user)}, nil
}

func (s *UserServiceServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	logger.Info(ctx, "GetUserById called:", req.Id)
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		logger.Error(ctx, "invalid user id:", err)
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.authUseCase.GetUserById(ctx, userId)
	if err != nil {
		logger.Error(ctx, "failed to get user by id:", err)
		return nil, err
	}
	return &pb.GetUserByIdResponse{User: dto.MapUserToUserDTO(&user)}, nil
}

func (s *UserServiceServer) LookupUserSession(ctx context.Context, req *pb.LookupUserSessionRequest) (*pb.LookupUserSessionResponse, error) {
	logger.Info(ctx, "LookupUserSession called:", req.SessionId)
	sessionId, err := uuid.Parse(req.SessionId)
	if err != nil {
		logger.Error(ctx, "invalid session id:", err)
		return nil, status.Error(codes.InvalidArgument, "invalid session id")
	}

	session := shared_models.Session{SessionId: sessionId}
	user, err := s.authUseCase.LookupUserSession(ctx, session)
	if err != nil {
		logger.Error(ctx, "failed to lookup user session:", err)
		return nil, err
	}

	return &pb.LookupUserSessionResponse{
		UserId:   user.Id.String(),
		Username: user.Username,
	}, nil
}

func (s *UserServiceServer) SearchSimilarUser(ctx context.Context, req *pb.SearchSimilarUserRequest) (*pb.SearchSimilarUserResponse, error) {
	logger.Info(ctx, "SearchSimilarUser called:", req.ToSearch)
	users, err := s.authUseCase.SearchSimilarUser(ctx, req.ToSearch, uint(req.NumUsers))
	if err != nil {
		logger.Error(ctx, "failed to search similar user:", err)
		return nil, err
	}

	protoUsers := make([]*pb.PublicUserInfo, len(users))
	for i, user := range users {
		protoUsers[i] = userclient.MapPublicUserInfoToDTO(&user)
	}

	return &pb.SearchSimilarUserResponse{UsersInfo: protoUsers}, nil
}
