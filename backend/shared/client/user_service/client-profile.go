package userclient

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/shared/logger"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type ProfileClient struct {
	client pb.ProfileServiceClient
}

func NewProfileClient(conn *grpc.ClientConn) *ProfileClient {
	return &ProfileClient{
		client: pb.NewProfileServiceClient(conn),
	}
}

func (c *ProfileClient) CreateProfile(ctx context.Context, profile shared_models.Profile) (shared_models.Profile, error) {
	logger.Info(ctx, "Sending request to create profile: %v", profile)
	resp, err := c.client.CreateProfile(ctx, &pb.CreateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	})
	if err != nil {
		logger.Error(ctx, "Failed to create profile: %v", err)
		return shared_models.Profile{}, err
	}

	newProfile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		logger.Error(ctx, "Failed to convert to Profile: %v", err)
		return shared_models.Profile{}, err
	}

	return *newProfile, nil
}

func (c *ProfileClient) UpdateProfile(ctx context.Context, profile shared_models.Profile) (*shared_models.Profile, error) {
	logger.Info(ctx, "Sending request to update profile: %v", profile)
	resp, err := c.client.UpdateProfile(ctx, &pb.UpdateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	})
	if err != nil {
		logger.Error(ctx, "Failed to update profile: %v", err)
		return nil, err
	}
	return MapProfileDTOToProfile(resp.Profile)
}

func (c *ProfileClient) GetProfile(ctx context.Context, userID uuid.UUID) (shared_models.Profile, error) {
	logger.Info(ctx, "Sending request to get profile: %v", userID)
	resp, err := c.client.GetProfile(ctx, &pb.GetProfileRequest{
		UserId: userID.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get profile: %v", err)
		return shared_models.Profile{}, err
	}
	profile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		logger.Error(ctx, "Failed to convert to Profile: %v", err)
		return shared_models.Profile{}, err
	}

	return *profile, nil
}

func (c *ProfileClient) GetProfileByUsername(ctx context.Context, username string) (shared_models.Profile, error) {
	logger.Info(ctx, "Sending request to get profile by username: %v", username)
	resp, err := c.client.GetProfileByUsername(ctx, &pb.GetProfileByUsernameRequest{
		Username: username,
	})
	if err != nil {
		logger.Error(ctx, "Failed to get profile by username: %v", err)
		return shared_models.Profile{}, err
	}
	profile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		logger.Error(ctx, "Failed to convert to Profile: %v", err)
		return shared_models.Profile{}, err
	}

	return *profile, nil
}

func (c *ProfileClient) UpdateLastSeen(ctx context.Context, userID uuid.UUID) error {
	logger.Info(ctx, "Sending request to update last seen: %v", userID)
	_, err := c.client.UpdateLastSeen(ctx, &pb.UpdateLastSeenRequest{
		UserId: userID.String(),
	})
	return err
}

func (c *ProfileClient) GetPublicUserInfo(ctx context.Context, userID uuid.UUID) (shared_models.PublicUserInfo, error) {
	logger.Info(ctx, "Sending request to get public user info: %v", userID)
	resp, err := c.client.GetPublicUserInfo(ctx, &pb.GetPublicUserInfoRequest{
		UserId: userID.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get public user info: %v", err)
		return shared_models.PublicUserInfo{}, err
	}

	info, err := MapPublicUserInfoDTOToModel(resp.UserInfo)
	if err != nil {
		logger.Error(ctx, "Failed to convert to PublicUserInfo: %v", err)
		return shared_models.PublicUserInfo{}, err
	}
	return *info, err
}

func (c *ProfileClient) GetPublicUsersInfo(ctx context.Context, userIDs []uuid.UUID) ([]shared_models.PublicUserInfo, error) {
	userIds := make([]string, len(userIDs))
	for i, id := range userIDs {
		userIds[i] = id.String()
	}

	logger.Info(ctx, "Sending request to get public user info for multiple users: %v", userIds)
	resp, err := c.client.GetPublicUsersInfo(ctx, &pb.GetPublicUsersInfoRequest{
		UserIds: userIds,
	})
	if err != nil {
		logger.Error(ctx, "Failed to get public user info: %v", err)
		return nil, err
	}

	publicInfos := make([]shared_models.PublicUserInfo, len(resp.UsersInfo))
	for i, userInfo := range resp.UsersInfo {
		info, err := MapPublicUserInfoDTOToModel(userInfo)
		if err != nil {
			logger.Error(ctx, "Failed to convert to PublicUserInfo: %v", err)
			return nil, err
		}
		publicInfos[i] = *info
	}

	return publicInfos, nil
}
