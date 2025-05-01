package userclient

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

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
	resp, err := c.client.CreateProfile(ctx, &pb.CreateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	})
	if err != nil {
		return shared_models.Profile{}, unwrapGRPCError(err)
	}

	newProfile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		return shared_models.Profile{}, err
	}

	return *newProfile, nil
}

func (c *ProfileClient) UpdateProfile(ctx context.Context, profile shared_models.Profile) (*shared_models.Profile, error) {
	resp, err := c.client.UpdateProfile(ctx, &pb.UpdateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	})
	if err != nil {
		return nil, unwrapGRPCError(err)
	}
	return MapProfileDTOToProfile(resp.Profile)
}

func (c *ProfileClient) GetProfile(ctx context.Context, userID uuid.UUID) (shared_models.Profile, error) {
	resp, err := c.client.GetProfile(ctx, &pb.GetProfileRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return shared_models.Profile{}, unwrapGRPCError(err)
	}
	profile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		return shared_models.Profile{}, err
	}

	return *profile, nil
}

func (c *ProfileClient) GetProfileByUsername(ctx context.Context, username string) (shared_models.Profile, error) {
	resp, err := c.client.GetProfileByUsername(ctx, &pb.GetProfileByUsernameRequest{
		Username: username,
	})
	if err != nil {
		return shared_models.Profile{}, unwrapGRPCError(err)
	}
	profile, err := MapProfileDTOToProfile(resp.Profile)
	if err != nil {
		return shared_models.Profile{}, err
	}

	return *profile, nil
}

func (c *ProfileClient) UpdateLastSeen(ctx context.Context, userID uuid.UUID) error {
	_, err := c.client.UpdateLastSeen(ctx, &pb.UpdateLastSeenRequest{
		UserId: userID.String(),
	})
	return unwrapGRPCError(err)
}

func (c *ProfileClient) GetPublicUserInfo(ctx context.Context, userID uuid.UUID) (shared_models.PublicUserInfo, error) {
	resp, err := c.client.GetPublicUserInfo(ctx, &pb.GetPublicUserInfoRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return shared_models.PublicUserInfo{}, unwrapGRPCError(err)
	}

	info, err := MapPublicUserInfoDTOToModel(resp.UserInfo)
	if err != nil {
		return shared_models.PublicUserInfo{}, err
	}
	return *info, err
}

func (c *ProfileClient) GetPublicUsersInfo(ctx context.Context, userIDs []uuid.UUID) ([]shared_models.PublicUserInfo, error) {
	userIds := make([]string, len(userIDs))
	for i, id := range userIDs {
		userIds[i] = id.String()
	}

	resp, err := c.client.GetPublicUsersInfo(ctx, &pb.GetPublicUsersInfoRequest{
		UserIds: userIds,
	})
	if err != nil {
		return nil, unwrapGRPCError(err)
	}

	publicInfos := make([]shared_models.PublicUserInfo, len(resp.UsersInfo))
	for i, userInfo := range resp.UsersInfo {
		info, err := MapPublicUserInfoDTOToModel(userInfo)
		if err != nil {
			return nil, err
		}
		publicInfos[i] = *info
	}

	return publicInfos, nil
}
