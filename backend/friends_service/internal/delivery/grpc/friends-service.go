package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	dto "quickflow/shared/client/friends_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
)

type FriendsUseCase interface {
	GetFriendsInfo(ctx context.Context, userID string, limit string, offset string, reqType string) ([]models.FriendInfo, int, error)
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
	logger.Info(ctx, "Received GetFriendsInfo request")

	friendsInfos, friendsCount, err := f.friendsUseCase.GetFriendsInfo(ctx, in.UserId, in.Limit, in.Offset, in.ReqType)
	if err != nil {
		logger.Error(ctx, "GetFriendsInfo failed: ", err)
		return &pb.GetFriendsInfoResponse{}, err
	}

	logger.Info(ctx, "Successfully fetched friends info")
	return dto.FromModelFriendsInfoToGrpc(friendsInfos, friendsCount), nil
}

func (f *FriendsServiceServer) SendFriendRequest(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	logger.Info(ctx, "Received SendFriendRequest request")

	exists, err := f.friendsUseCase.IsExistsFriendRequest(ctx, in.UserId, in.ReceiverId)
	if err != nil {
		logger.Error(ctx, "IsExistsFriendRequest failed: ", err)
		return &emptypb.Empty{}, err
	}

	if exists {
		logger.Info(ctx, "Friend request already exists")
		return &emptypb.Empty{}, errors.New("friend request already exists")
	}

	if err := f.friendsUseCase.SendFriendRequest(ctx, in.UserId, in.ReceiverId); err != nil {
		logger.Error(ctx, "SendFriendRequest failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully sent friend request")
	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) AcceptFriendRequest(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	logger.Info(ctx, "Received AcceptFriendRequest request")

	if err := f.friendsUseCase.AcceptFriendRequest(ctx, in.UserId, in.ReceiverId); err != nil {
		logger.Error(ctx, "AcceptFriendRequest failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully accepted friend request")
	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) Unfollow(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	logger.Info(ctx, "Received Unfollow request")

	if err := f.friendsUseCase.Unfollow(ctx, in.UserId, in.ReceiverId); err != nil {
		logger.Error(ctx, "Unfollow failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully unfollowed user")
	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) DeleteFriend(ctx context.Context, in *pb.FriendRequest) (*emptypb.Empty, error) {
	logger.Info(ctx, "Received DeleteFriend request")

	if err := f.friendsUseCase.DeleteFriend(ctx, in.UserId, in.ReceiverId); err != nil {
		logger.Error(ctx, "DeleteFriend failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully deleted friend")
	return &emptypb.Empty{}, nil
}

func (f *FriendsServiceServer) GetUserRelation(ctx context.Context, in *pb.FriendRequest) (*pb.RelationResponse, error) {
	logger.Info(ctx, "Received GetUserRelation request")

	user1Id, err := uuid.Parse(in.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: ", err)
		return &pb.RelationResponse{}, nil
	}
	user2Id, err := uuid.Parse(in.ReceiverId)
	if err != nil {
		logger.Error(ctx, "Invalid ReceiverId: ", err)
		return &pb.RelationResponse{}, nil
	}

	rel, err := f.friendsUseCase.GetUserRelation(ctx, user1Id, user2Id)
	if err != nil {
		logger.Error(ctx, "GetUserRelation failed: ", err)
		return &pb.RelationResponse{}, err
	}

	logger.Info(ctx, "Successfully fetched user relation")
	return &pb.RelationResponse{
		Relation: string(rel),
	}, nil
}
