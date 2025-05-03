package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/shared/logger"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
	"quickflow/user_service/internal/delivery/grpc/dto"
	user_errors "quickflow/user_service/internal/errors"
)

type ProfileUseCase interface {
	CreateProfile(ctx context.Context, profile shared_models.Profile) (shared_models.Profile, error)
	UpdateProfile(ctx context.Context, profile shared_models.Profile) (*shared_models.Profile, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (shared_models.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (shared_models.Profile, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID) error
	GetPublicUserInfo(ctx context.Context, userID uuid.UUID) (shared_models.PublicUserInfo, error)
	GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]shared_models.PublicUserInfo, error)
}

type ProfileServiceServer struct {
	pb.UnimplementedProfileServiceServer
	profileUC ProfileUseCase
}

func NewProfileServiceServer(profileUC ProfileUseCase) *ProfileServiceServer {
	return &ProfileServiceServer{profileUC: profileUC}
}

func (p *ProfileServiceServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	logger.Info(ctx, "CreateProfile called")

	profileDTO := req.GetProfile()
	if profileDTO == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := dto.MapProfileDTOToProfile(profileDTO)
	if err != nil {
		logger.Error(ctx, "invalid profile data:", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid profile data: %v", err)
	}

	createdProfile, err := p.profileUC.CreateProfile(ctx, *profile)
	if err != nil {
		logger.Error(ctx, "failed to create profile:", err)
		return nil, status.Errorf(codes.Internal, "failed to create profile: %v", err)
	}

	return &pb.CreateProfileResponse{
		Profile: dto.MapProfileToProfileDTO(&createdProfile),
	}, nil
}

func (p *ProfileServiceServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	logger.Info(ctx, "UpdateProfile called")

	profileDTO := req.GetProfile()
	if profileDTO == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := dto.MapProfileDTOToProfile(profileDTO)
	if err != nil {
		logger.Error(ctx, "invalid profile data:", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid profile data: %v", err)
	}

	updatedProfile, err := p.profileUC.UpdateProfile(ctx, *profile)
	if errors.Is(err, user_errors.ErrInvalidProfileInfo) {
		return nil, status.Error(codes.InvalidArgument, "invalid profile info")
	}
	if errors.Is(err, user_errors.ErrAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	}
	if err != nil {
		logger.Error(ctx, "failed to update profile:", err)
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	return &pb.UpdateProfileResponse{
		Profile: dto.MapProfileToProfileDTO(updatedProfile),
	}, nil
}

func (p *ProfileServiceServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	logger.Info(ctx, "GetProfile called")

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	profile, err := p.profileUC.GetProfile(ctx, userID)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	if err != nil {
		logger.Error(ctx, "failed to get profile:", err)
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &pb.GetProfileResponse{
		Profile: dto.MapProfileToProfileDTO(&profile),
	}, nil
}

func (p *ProfileServiceServer) GetProfileByUsername(ctx context.Context, req *pb.GetProfileByUsernameRequest) (*pb.GetProfileByUsernameResponse, error) {
	logger.Info(ctx, "GetProfileByUsername called")

	username := req.Username
	if len(username) == 0 {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	profile, err := p.profileUC.GetProfileByUsername(ctx, username)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	if err != nil {
		logger.Error(ctx, "failed to get profile by username:", err)
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &pb.GetProfileByUsernameResponse{
		Profile: dto.MapProfileToProfileDTO(&profile),
	}, nil
}

func (p *ProfileServiceServer) UpdateLastSeen(ctx context.Context, req *pb.UpdateLastSeenRequest) (*pb.UpdateLastSeenResponse, error) {
	logger.Info(ctx, "UpdateLastSeen called")

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := p.profileUC.UpdateLastSeen(ctx, userID); err != nil {
		logger.Error(ctx, "failed to update last seen:", err)
		return nil, status.Errorf(codes.Internal, "failed to update last seen: %v", err)
	}

	return &pb.UpdateLastSeenResponse{Success: true}, nil
}

func (p *ProfileServiceServer) GetPublicUserInfo(ctx context.Context, req *pb.GetPublicUserInfoRequest) (*pb.GetPublicUserInfoResponse, error) {
	logger.Info(ctx, "GetPublicUserInfo called")

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	info, err := p.profileUC.GetPublicUserInfo(ctx, userID)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		logger.Error(ctx, "failed to get public user info:", err)
		return nil, status.Errorf(codes.Internal, "failed to get public user info: %v", err)
	}

	return &pb.GetPublicUserInfoResponse{
		UserInfo: dto.MapPublicUserInfoToDTO(&info),
	}, nil
}

func (p *ProfileServiceServer) GetPublicUsersInfo(ctx context.Context, req *pb.GetPublicUsersInfoRequest) (*pb.GetPublicUsersInfoResponse, error) {
	logger.Info(ctx, "GetPublicUsersInfo called")

	userIds := req.GetUserIds()
	if len(userIds) == 0 {
		return &pb.GetPublicUsersInfoResponse{
			UsersInfo: nil,
		}, nil
	}

	parsedUserIds := make([]uuid.UUID, len(userIds))
	for i, id := range userIds {
		parsedId, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid user id")
		}
		parsedUserIds[i] = parsedId
	}

	publicInfo, err := p.profileUC.GetPublicUsersInfo(ctx, parsedUserIds)
	if err != nil {
		logger.Error(ctx, "failed to get public users info:", err)
		return nil, status.Errorf(codes.Internal, "failed to get public users info: %v", err)
	}

	var userInfos []*pb.PublicUserInfo
	for _, info := range publicInfo {
		userInfos = append(userInfos, dto.MapPublicUserInfoToDTO(&info))
	}

	return &pb.GetPublicUsersInfoResponse{
		UsersInfo: userInfos,
	}, nil
}
