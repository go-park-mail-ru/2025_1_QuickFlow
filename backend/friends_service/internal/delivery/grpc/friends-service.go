package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"quickflow/friends_service/internal/delivery/dto"
	pb "quickflow/friends_service/internal/delivery/grpc/proto"
	"quickflow/shared/models"
)

type FriendsUseCase interface {
	GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]models.FriendInfo, int, error)
	SendFriendRequest(ctx context.Context, senderID string, receiverID string) error
	AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error
	Unfollow(ctx context.Context, userID string, friendID string) error
	DeleteFriend(ctx context.Context, user string, friend string) error
	IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error)
	GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error)
}

type FriendsServiceServer struct {
	pb.UnimplementedFriendsServiceServer
	friendsUseCase FriendsUseCase
}

func NewFriendsServiceServer(friendsUseCase FriendsUseCase) *FriendsServiceServer {
	return &FriendsServiceServer{
		friendsUseCase: friendsUseCase,
	}
}

func (f *FriendsServiceServer) GetFriendsInfo(ctx context.Context, in *pb.GetFriendsInfoRequest) (*pb.GetFriendsInfoResponse, error) {
	friendsInfos, friendsCount, err := f.friendsUseCase.GetFriendsInfo(ctx, in.UserId, in.Limit, in.Offset)
	if err != nil {
		return &pb.GetFriendsInfoResponse{}, err
	}

	return dto.FromModelFriendsInfoToGrpc(friendsInfos, friendsCount), nil
}
func (f *FriendsServiceServer) SendFriendRequest(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	if err := f.friendsUseCase.SendFriendRequest(ctx, in.UserId, in.ReceiverId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) AcceptFriendRequest(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	if err := f.friendsUseCase.AcceptFriendRequest(ctx, in.UserId, in.ReceiverId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) Unfollow(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	if err := f.friendsUseCase.Unfollow(ctx, in.UserId, in.ReceiverId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) DeleteFriend(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	if err := f.friendsUseCase.DeleteFriend(ctx, in.UserId, in.ReceiverId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) GetUserRelation(ctx context.Context, in *pb.FriendRequest) (*pb.RelationResponse, error) {
	user1Id, err := uuid.Parse(in.UserId)
	if err != nil {
		return &pb.RelationResponse{}, nil
	}
	user2Id, err := uuid.Parse(in.ReceiverId)
	if err != nil {
		return &pb.RelationResponse{}, nil
	}

	rel, err := f.friendsUseCase.GetUserRelation(ctx, user1Id, user2Id)
	if err != nil {
		return &pb.RelationResponse{}, err
	}

	return &pb.RelationResponse{
		Relation: string(rel),
	}, nil
}
