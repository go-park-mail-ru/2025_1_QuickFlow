package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shared_models "quickflow/shared/models"
	"quickflow/user_service/internal/delivery/grpc/dto"
	pb "quickflow/user_service/internal/delivery/grpc/proto"
	user_errors "quickflow/user_service/internal/errors"
)

type ProfileUseCase interface {
	CreateProfile(ctx context.Context, profile shared_models.Profile) (shared_models.Profile, error)
	UpdateProfile(ctx context.Context, profile shared_models.Profile) (*shared_models.Profile, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (shared_models.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (shared_models.Profile, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID) error
	GetPublicUserInfo(ctx context.Context, userID uuid.UUID) (shared_models.PublicUserInfo, error)
}

type ProfileServiceServer struct {
	pb.UnimplementedProfileServiceServer
	profileUC ProfileUseCase
}

func NewProfileServiceServer(profileUC ProfileUseCase) *ProfileServiceServer {
	return &ProfileServiceServer{profileUC: profileUC}
}

// CreateProfile создает новый профиль
func (p *ProfileServiceServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	profileDTO := req.GetProfile()
	if profileDTO == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := dto.MapProfileDTOToProfile(profileDTO)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid profile data: %v", err)
	}

	createdProfile, err := p.profileUC.CreateProfile(ctx, *profile)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create profile: %v", err)
	}

	return &pb.CreateProfileResponse{
		Profile: dto.MapProfileToProfileDTO(&createdProfile),
	}, nil
}

// UpdateProfile обновляет существующий профиль
func (p *ProfileServiceServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	profileDTO := req.GetProfile()
	if profileDTO == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := dto.MapProfileDTOToProfile(profileDTO)
	if err != nil {
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
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	return &pb.UpdateProfileResponse{
		Profile: dto.MapProfileToProfileDTO(updatedProfile),
	}, nil
}

// GetProfile получает профиль пользователя по user_id
func (p *ProfileServiceServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	profile, err := p.profileUC.GetProfile(ctx, userID)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &pb.GetProfileResponse{
		Profile: dto.MapProfileToProfileDTO(&profile),
	}, nil
}

// GetProfileByUsername получает профиль пользователя по username
func (p *ProfileServiceServer) GetProfileByUsername(ctx context.Context, req *pb.GetProfileByUsernameRequest) (*pb.GetProfileByUsernameResponse, error) {
	username := req.GetUsername()
	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	profile, err := p.profileUC.GetProfileByUsername(ctx, username)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &pb.GetProfileByUsernameResponse{
		Profile: dto.MapProfileToProfileDTO(&profile),
	}, nil
}

// UpdateLastSeen обновляет время последней активности пользователя
func (p *ProfileServiceServer) UpdateLastSeen(ctx context.Context, req *pb.UpdateLastSeenRequest) (*pb.UpdateLastSeenResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := p.profileUC.UpdateLastSeen(ctx, userID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update last seen: %v", err)
	}

	return &pb.UpdateLastSeenResponse{Success: true}, nil
}

// GetPublicUserInfo получает публичную информацию о пользователе
func (p *ProfileServiceServer) GetPublicUserInfo(ctx context.Context, req *pb.GetPublicUserInfoRequest) (*pb.GetPublicUserInfoResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	info, err := p.profileUC.GetPublicUserInfo(ctx, userID)
	if errors.Is(err, user_errors.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get public user info: %v", err)
	}

	return &pb.GetPublicUserInfoResponse{
		UserInfo: dto.MapPublicUserInfoToDTO(&info),
	}, nil
}
