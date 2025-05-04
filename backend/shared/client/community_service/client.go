package community_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"
)

type Client struct {
	client pb.CommunityServiceClient
}

func NewCommunityServiceClient(conn *grpc.ClientConn) *Client {
	return &Client{
		client: pb.NewCommunityServiceClient(conn),
	}
}

func (c *Client) CreateCommunity(ctx context.Context, community *models.Community) (*models.Community, error) {
	res, err := c.client.CreateCommunity(ctx, &pb.CreateCommunityRequest{
		Name:        community.Name,
		Description: community.Description,
		AvatarUrl:   community.AvatarUrl,
		Avatar:      file_service.ModelFileToProto(community.Avatar),
		OwnerId:     community.OwnerID.String(),
	})
	if err != nil {
		return nil, err
	}
	return MapProtoCommunityToModel(res.Community)
}

func (c *Client) GetCommunityById(ctx context.Context, id uuid.UUID) (*models.Community, error) {
	res, err := c.client.GetCommunityById(ctx, &pb.GetCommunityByIdRequest{CommunityId: id.String()})
	if err != nil {
		return nil, err
	}
	return MapProtoCommunityToModel(res.Community)
}

func (c *Client) GetCommunityByName(ctx context.Context, name string) (*models.Community, error) {
	res, err := c.client.GetCommunityByName(ctx, &pb.GetCommunityByNameRequest{CommunityName: name})
	if err != nil {
		return nil, err
	}
	return MapProtoCommunityToModel(res.Community)
}

func (c *Client) IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error) {
	res, err := c.client.IsCommunityMember(ctx, &pb.IsCommunityMemberRequest{
		UserId:      userId.String(),
		CommunityId: communityId.String(),
	})
	if err != nil {
		return false, nil, err
	}

	if res.Role == -1 {
		return res.IsMember, nil, nil
	}
	role := convertRoleFromProto(res.Role)
	return res.IsMember, &role, nil
}

func (c *Client) GetCommunityMembers(ctx context.Context, communityId uuid.UUID, count int, ts time.Time) ([]*models.CommunityMember, error) {
	res, err := c.client.GetCommunityMembers(ctx, &pb.GetCommunityMembersRequest{
		CommunityId: communityId.String(),
		Count:       int32(count),
		Ts:          timestamppb.New(ts),
	})
	if err != nil {
		return nil, err
	}
	var result []*models.CommunityMember
	for _, protoMember := range res.Members {
		m, err := MapProtoMemberToModel(protoMember)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (c *Client) DeleteCommunity(ctx context.Context, communityId uuid.UUID, userId uuid.UUID) error {
	_, err := c.client.DeleteCommunity(ctx, &pb.DeleteCommunityRequest{
		CommunityId: communityId.String(),
		UserId:      userId.String(),
	})
	return err
}

func (c *Client) UpdateCommunity(ctx context.Context, community *models.Community, userId uuid.UUID) (*models.Community, error) {
	res, err := c.client.UpdateCommunity(ctx, &pb.UpdateCommunityRequest{
		Id:          community.ID.String(),
		Name:        community.Name,
		Description: community.Description,
		AvatarUrl:   community.AvatarUrl,
		Avatar:      file_service.ModelFileToProto(community.Avatar),
		UserId:      userId.String(),
	})
	if err != nil {
		return nil, err
	}
	return MapProtoCommunityToModel(res.Community)
}

func (c *Client) JoinCommunity(ctx context.Context, member *models.CommunityMember) error {
	_, err := c.client.JoinCommunity(ctx, &pb.JoinCommunityRequest{
		NewMember: MapModelMemberToProto(member),
	})
	return err
}

func (c *Client) LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error {
	_, err := c.client.LeaveCommunity(ctx, &pb.LeaveCommunityRequest{
		CommunityId: communityId.String(),
		UserId:      userId.String(),
	})
	return err
}

func (c *Client) GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]*models.Community, error) {
	res, err := c.client.GetUserCommunities(ctx, &pb.GetUserCommunitiesRequest{
		UserId: userId.String(),
		Count:  int32(count),
		Ts:     timestamppb.New(ts),
	})
	if err != nil {
		return nil, err
	}
	var result []*models.Community
	for _, protoComm := range res.Communities {
		m, err := MapProtoCommunityToModel(protoComm)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (c *Client) SearchSimilarCommunities(ctx context.Context, name string, count int) ([]*models.Community, error) {
	res, err := c.client.SearchSimilarCommunities(ctx, &pb.SearchSimilarCommunitiesRequest{
		Name:  name,
		Count: int32(count),
	})
	if err != nil {
		return nil, err
	}
	var result []*models.Community
	for _, protoComm := range res.Communities {
		m, err := MapProtoCommunityToModel(protoComm)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}
