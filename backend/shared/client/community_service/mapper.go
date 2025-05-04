package community_service

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"
)

func MapContactInfoToDTO(contactInfo *models.ContactInfo) *pb.ContactInfo {
	if contactInfo == nil {
		return nil
	}

	return &pb.ContactInfo{
		Email:       contactInfo.Email,
		PhoneNumber: contactInfo.Phone,
		City:        contactInfo.City,
	}
}

func MapContactInfoDTOToModel(contactInfoDTO *pb.ContactInfo) *models.ContactInfo {
	if contactInfoDTO == nil {
		return nil
	}

	return &models.ContactInfo{
		Email: contactInfoDTO.Email,
		Phone: contactInfoDTO.PhoneNumber,
		City:  contactInfoDTO.City,
	}
}

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
		ID:        id,
		NickName:  community.Nickname,
		OwnerID:   ownerId,
		CreatedAt: community.CreatedAt.AsTime(),
		Avatar:    file_service.ProtoFileToModel(community.Avatar),
		Cover:     file_service.ProtoFileToModel(community.Cover),
		BasicInfo: &models.BasicCommunityInfo{
			Name:        community.Name,
			Description: community.Description,
			AvatarUrl:   community.AvatarUrl,
			CoverUrl:    community.CoverUrl,
		},
		ContactInfo: MapContactInfoDTOToModel(community.ContactInfo),
	}, nil

}

func MapModelCommunityToProto(community *models.Community) *pb.Community {
	if community == nil {
		return nil
	}

	return &pb.Community{
		Id:          community.ID.String(),
		OwnerId:     community.OwnerID.String(),
		Name:        community.BasicInfo.Name,
		Description: community.BasicInfo.Description,
		CreatedAt:   timestamppb.New(community.CreatedAt),
		AvatarUrl:   community.BasicInfo.AvatarUrl,
		CoverUrl:    community.BasicInfo.CoverUrl,
		Nickname:    community.NickName,
		Avatar:      file_service.ModelFileToProto(community.Avatar),
		Cover:       file_service.ModelFileToProto(community.Cover),
		ContactInfo: MapContactInfoToDTO(community.ContactInfo),
	}

}

func MapModelMemberToProto(member *models.CommunityMember) *pb.CommunityMember {
	if member == nil {
		return nil
	}

	return &pb.CommunityMember{
		UserId:      member.UserID.String(),
		CommunityId: member.CommunityID.String(),
		Role:        ConvertRoleToProto(member.Role),
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
		Role:        ConvertRoleFromProto(member.Role),
		JoinedAt:    member.JoinedAt.AsTime(),
	}, nil
}

func ConvertRoleToProto(role models.CommunityRole) pb.CommunityRole {
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

func ConvertRoleFromProto(role pb.CommunityRole) models.CommunityRole {
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
