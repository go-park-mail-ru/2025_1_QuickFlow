package community_service

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"
)

func MapProtoCommunityToModel(community *pb.Community) (*models.Community, error) {
	if community == nil {
		return nil, nil
	}

	id, err := uuid.Parse(community.Id)
	if err != nil {
		return nil, err
	}

	ownerId, err := uuid.Parse(community.OwnerId)
	if err != nil {
		return nil, err
	}

	return &models.Community{
		ID:          id,
		OwnerID:     ownerId,
		Name:        community.Name,
		Description: community.Description,
		CreatedAt:   community.CreatedAt.AsTime(),
		AvatarUrl:   community.AvatarUrl,
		Avatar:      file_service.ProtoFileToModel(community.Avatar),
	}, nil
}

func MapModelCommunityToProto(community *models.Community) *pb.Community {

	return &pb.Community{
		Id:          community.ID.String(),
		OwnerId:     community.OwnerID.String(),
		Name:        community.Name,
		Description: community.Description,
		CreatedAt:   timestamppb.New(community.CreatedAt),
		AvatarUrl:   community.AvatarUrl,
	}
}

func MapModelMemberToProto(member *models.CommunityMember) *pb.CommunityMember {
	if member == nil {
		return nil
	}

	return &pb.CommunityMember{
		UserId:      member.UserID.String(),
		CommunityId: member.CommunityID.String(),
		Role:        convertRoleToProto(member.Role),
		JoinedAt:    timestamppb.New(member.JoinedAt),
	}
}

func MapProtoMemberToModel(member *pb.CommunityMember) (*models.CommunityMember, error) {
	if member == nil {
		return nil, nil
	}

	userId, err := uuid.Parse(member.UserId)
	if err != nil {
		return nil, err
	}

	communityId, err := uuid.Parse(member.CommunityId)
	if err != nil {
		return nil, err
	}

	return &models.CommunityMember{
		UserID:      userId,
		CommunityID: communityId,
		Role:        convertRoleFromProto(member.Role),
		JoinedAt:    member.JoinedAt.AsTime(),
	}, nil
}

func convertRoleToProto(role models.CommunityRole) pb.CommunityRole {
	switch role {
	case models.CommunityRoleMember:
		return pb.CommunityRole_COMMUNITY_ROLE_MEMBER
	case models.CommunityRoleAdmin:
		return pb.CommunityRole_COMMUNITY_ROLE_ADMIN
	case models.CommunityRoleOwner:
		return pb.CommunityRole_COMMUNITY_ROLE_OWNER
	default:
		return pb.CommunityRole_COMMUNITY_ROLE_MEMBER
	}
}

func convertRoleFromProto(role pb.CommunityRole) models.CommunityRole {
	switch role {
	case pb.CommunityRole_COMMUNITY_ROLE_MEMBER:
		return models.CommunityRoleMember
	case pb.CommunityRole_COMMUNITY_ROLE_ADMIN:
		return models.CommunityRoleAdmin
	case pb.CommunityRole_COMMUNITY_ROLE_OWNER:
		return models.CommunityRoleOwner
	default:
		return models.CommunityRoleMember
	}
}
