package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dto "quickflow/shared/client/community_service"
	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"
)

type CommunityUseCase interface {
	CreateCommunity(ctx context.Context, community models.Community) (*models.Community, error)
	GetCommunityById(ctx context.Context, id uuid.UUID) (models.Community, error)
	GetCommunityByName(ctx context.Context, name string) (models.Community, error)
	GetCommunityMembers(ctx context.Context, id uuid.UUID, numMembers int, ts time.Time) ([]models.CommunityMember, error)
	IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error)
	DeleteCommunity(ctx context.Context, id uuid.UUID) error
	UpdateCommunity(ctx context.Context, community models.Community, userId uuid.UUID) (*models.Community, error)
	JoinCommunity(ctx context.Context, member models.CommunityMember) error
	LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error
	GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error)
	SearchSimilarCommunities(ctx context.Context, name string, count int) ([]models.Community, error)
	ChangeUserRole(ctx context.Context, userId, communityId uuid.UUID, role models.CommunityRole, requester uuid.UUID) error
	GetControlledCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error)
}

type UserUseCase interface {
	GetUserById(ctx context.Context, userId uuid.UUID) (*shared_models.User, error)
}

type CommunityServiceServer struct {
	pb.UnimplementedCommunityServiceServer
	communityUseCase CommunityUseCase
}

func NewCommunityServiceServer(communityUseCase CommunityUseCase) *CommunityServiceServer {
	return &CommunityServiceServer{
		communityUseCase: communityUseCase,
	}
}

func (s *CommunityServiceServer) CreateCommunity(ctx context.Context, req *pb.CreateCommunityRequest) (*pb.CreateCommunityResponse, error) {
	ownerId, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner ID")
	}

	community := models.Community{
		NickName: req.Nickname,
		OwnerID:  ownerId,
		Avatar:   file_service.ProtoFileToModel(req.Avatar),
		Cover:    file_service.ProtoFileToModel(req.Cover),
		BasicInfo: &models.BasicCommunityInfo{
			Name:        req.Name,
			Description: req.Description,
			CoverUrl:    req.CoverUrl,
			AvatarUrl:   req.AvatarUrl,
		},
	}

	newCommunity, err := s.communityUseCase.CreateCommunity(ctx, community)
	if err != nil {
		return nil, err
	}

	return &pb.CreateCommunityResponse{
		Community: dto.MapModelCommunityToProto(newCommunity),
	}, nil
}

func (s *CommunityServiceServer) GetCommunityById(ctx context.Context, req *pb.GetCommunityByIdRequest) (*pb.GetCommunityByIdResponse, error) {
	id, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	community, err := s.communityUseCase.GetCommunityById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetCommunityByIdResponse{
		Community: dto.MapModelCommunityToProto(&community),
	}, nil
}

func (s *CommunityServiceServer) GetCommunityByName(ctx context.Context, req *pb.GetCommunityByNameRequest) (*pb.GetCommunityByNameResponse, error) {
	community, err := s.communityUseCase.GetCommunityByName(ctx, req.CommunityName)
	if err != nil {
		return nil, err
	}

	return &pb.GetCommunityByNameResponse{
		Community: dto.MapModelCommunityToProto(&community),
	}, nil
}

func (s *CommunityServiceServer) GetCommunityMembers(ctx context.Context, req *pb.GetCommunityMembersRequest) (*pb.GetCommunityMembersResponse, error) {
	id, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	members, err := s.communityUseCase.GetCommunityMembers(ctx, id, int(req.Count), req.Ts.AsTime())
	if err != nil {
		return nil, err
	}

	protoMembers := make([]*pb.CommunityMember, len(members))
	for i, member := range members {
		protoMembers[i] = dto.MapModelMemberToProto(&member)
	}

	return &pb.GetCommunityMembersResponse{
		Members: protoMembers,
	}, nil
}

func (s *CommunityServiceServer) IsCommunityMember(ctx context.Context, req *pb.IsCommunityMemberRequest) (*pb.IsCommunityMemberResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	communityId, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	isMember, role, err := s.communityUseCase.IsCommunityMember(ctx, userId, communityId)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return &pb.IsCommunityMemberResponse{
			IsMember: isMember,
			Role:     -1,
		}, nil
	}
	return &pb.IsCommunityMemberResponse{
		IsMember: isMember,
		Role:     dto.ConvertRoleToProto(*role),
	}, nil
}

func (s *CommunityServiceServer) DeleteCommunity(ctx context.Context, req *pb.DeleteCommunityRequest) (*pb.DeleteCommunityResponse, error) {
	id, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	err = s.communityUseCase.DeleteCommunity(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteCommunityResponse{}, nil
}

func (s *CommunityServiceServer) UpdateCommunity(ctx context.Context, req *pb.UpdateCommunityRequest) (*pb.UpdateCommunityResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	updatedCommunity, err := s.communityUseCase.UpdateCommunity(ctx, models.Community{
		ID:       id,
		OwnerID:  userId,
		NickName: req.Nickname,
		Avatar:   file_service.ProtoFileToModel(req.Avatar),
		Cover:    file_service.ProtoFileToModel(req.Cover),
		BasicInfo: &models.BasicCommunityInfo{
			Name:        req.Name,
			Description: req.Description,
			CoverUrl:    req.CoverUrl,
			AvatarUrl:   req.AvatarUrl,
		},
		ContactInfo: dto.MapContactInfoDTOToModel(req.ContactInfo),
	}, userId)

	if err != nil {
		return nil, err
	}

	return &pb.UpdateCommunityResponse{
		Community: dto.MapModelCommunityToProto(updatedCommunity),
	}, nil
}

func (s *CommunityServiceServer) JoinCommunity(ctx context.Context, req *pb.JoinCommunityRequest) (*pb.JoinCommunityResponse, error) {
	member, err := dto.MapProtoMemberToModel(req.NewMember)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid member data")
	}
	if member == nil {
		return nil, status.Error(codes.InvalidArgument, "member data cannot be empty")
	}

	err = s.communityUseCase.JoinCommunity(ctx, *member)
	if err != nil {
		return nil, err
	}

	return &pb.JoinCommunityResponse{
		Success: true,
	}, nil
}

func (s *CommunityServiceServer) LeaveCommunity(ctx context.Context, req *pb.LeaveCommunityRequest) (*pb.LeaveCommunityResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	communityId, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	err = s.communityUseCase.LeaveCommunity(ctx, userId, communityId)
	if err != nil {
		return nil, err
	}

	return &pb.LeaveCommunityResponse{
		Success: true,
	}, nil
}

func (s *CommunityServiceServer) GetUserCommunities(ctx context.Context, req *pb.GetUserCommunitiesRequest) (*pb.GetUserCommunitiesResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	communities, err := s.communityUseCase.GetUserCommunities(ctx, userId, int(req.Count), req.Ts.AsTime())
	if err != nil {
		return nil, err
	}

	protoCommunities := make([]*pb.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = dto.MapModelCommunityToProto(&community)
	}

	return &pb.GetUserCommunitiesResponse{
		Communities: protoCommunities,
	}, nil
}

func (s *CommunityServiceServer) SearchSimilarCommunities(ctx context.Context, req *pb.SearchSimilarCommunitiesRequest) (*pb.SearchSimilarCommunitiesResponse, error) {
	communities, err := s.communityUseCase.SearchSimilarCommunities(ctx, req.Name, int(req.Count))
	if err != nil {
		return nil, err
	}

	protoCommunities := make([]*pb.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = dto.MapModelCommunityToProto(&community)
	}

	return &pb.SearchSimilarCommunitiesResponse{
		Communities: protoCommunities,
	}, nil
}

func (s *CommunityServiceServer) ChangeUserRole(ctx context.Context, req *pb.ChangeUserRoleRequest) (*pb.ChangeUserRoleResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	communityId, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	requesterId, err := uuid.Parse(req.RequesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid requester ID")
	}

	role := dto.ConvertRoleFromProto(req.Role)

	err = s.communityUseCase.ChangeUserRole(ctx, userId, communityId, role, requesterId)
	if err != nil {
		return nil, err
	}

	return &pb.ChangeUserRoleResponse{
		Success: true,
	}, nil
}

func (s *CommunityServiceServer) GetControlledCommunities(ctx context.Context, req *pb.GetControlledCommunitiesRequest) (*pb.GetControlledCommunitiesResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	communities, err := s.communityUseCase.GetControlledCommunities(ctx, userId, int(req.Count), req.Ts.AsTime())
	if err != nil {
		return nil, err
	}

	protoCommunities := make([]*pb.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = dto.MapModelCommunityToProto(&community)
	}

	return &pb.GetControlledCommunitiesResponse{
		Communities: protoCommunities,
	}, nil
}
