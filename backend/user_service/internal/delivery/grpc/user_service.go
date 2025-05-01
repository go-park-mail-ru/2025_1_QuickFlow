package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
	"quickflow/user_service/internal/delivery/grpc/dto"
	user_errors "quickflow/user_service/internal/errors"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, user shared_models.User, profile shared_models.Profile) (uuid.UUID, shared_models.Session, error)
	AuthUser(ctx context.Context, authData shared_models.LoginData) (shared_models.Session, error)
	GetUserByUsername(ctx context.Context, username string) (shared_models.User, error)
	LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error)
	DeleteUserSession(ctx context.Context, session string) error
	GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error)
}
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	authUseCase UserUseCase
}

func NewUserServiceServer(authUseCase UserUseCase) *UserServiceServer {
	return &UserServiceServer{authUseCase: authUseCase}
}

func (s *UserServiceServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	user, err := dto.MapUserDTOToUser(req.User)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("invalid user data: %w", err))
	}

	profile, err := dto.MapProfileDTOToProfile(req.Profile)
	if err != nil {
		return nil, grpcErrorFromAppError(fmt.Errorf("invalid profile data: %w", err))
	}

	_, session, err := s.authUseCase.CreateUser(ctx, *user, *profile)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.SignUpResponse{
		Session: dto.MapSessionToDTO(session),
	}, nil
}

func (s *UserServiceServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	loginData := dto.MapSignInToSignInDTO(req.SignIn)

	session, err := s.authUseCase.AuthUser(ctx, *loginData)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.SignInResponse{
		Session: dto.MapSessionToDTO(session),
	}, nil
}

func (s *UserServiceServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	err := s.authUseCase.DeleteUserSession(ctx, req.SessionId)
	if err != nil {
		return &pb.SignOutResponse{Success: false}, grpcErrorFromAppError(err)
	}
	return &pb.SignOutResponse{Success: true}, nil
}

func (s *UserServiceServer) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserByUsernameResponse, error) {
	user, err := s.authUseCase.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetUserByUsernameResponse{
		User: dto.MapUserToUserDTO(&user),
	}, nil
}

func (s *UserServiceServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.authUseCase.GetUserById(ctx, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetUserByIdResponse{
		User: dto.MapUserToUserDTO(&user),
	}, nil
}

func (s *UserServiceServer) LookupUserSession(ctx context.Context, req *pb.LookupUserSessionRequest) (*pb.LookupUserSessionResponse, error) {
	sessionId, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session id")
	}

	session := shared_models.Session{SessionId: sessionId}
	user, err := s.authUseCase.LookupUserSession(ctx, session)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.LookupUserSessionResponse{
		UserId:   user.Id.String(),
		Username: user.Username,
	}, nil
}

func grpcErrorFromAppError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, user_errors.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, user_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
