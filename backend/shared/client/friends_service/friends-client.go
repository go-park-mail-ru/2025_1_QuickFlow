package friends_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/shared/logger"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
)

type FriendsClient struct {
	client pb.FriendsServiceClient
}

func NewFriendsClient(conn *grpc.ClientConn) *FriendsClient {
	return &FriendsClient{
		client: pb.NewFriendsServiceClient(conn),
	}
}

func (f *FriendsClient) GetFriendsInfo(ctx context.Context, userID string, limit string, offset string, reqType string) ([]shared_models.FriendInfo, int, error) {
	logger.Info(ctx, fmt.Sprintf("Getting friends info for userId: %s", userID))
	info, err := f.client.GetFriendsInfo(ctx, &pb.GetFriendsInfoRequest{
		UserId:  userID,
		Limit:   limit,
		Offset:  offset,
		ReqType: reqType,
	})

	if err != nil {
		logger.Error(ctx, "Failed to get friends info: %v", err)
		return nil, 0, err
	}

	res, friendsCount := FromGrpcToModelFriendsInfo(info)

	return res, friendsCount, nil
}

func (f *FriendsClient) GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (shared_models.UserRelation, error) {
	logger.Info(ctx, fmt.Sprintf("Getting relation between %v and %v", user1, user2))
	relation, err := f.client.GetUserRelation(ctx, &pb.FriendRequest{UserId: user1.String(), ReceiverId: user2.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get user relation: %v", err)
		return "", err
	}

	return shared_models.UserRelation(relation.Relation), nil
}

func (f *FriendsClient) SendFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	logger.Info(ctx, fmt.Sprintf("Sending friend request to %v from %v", receiverID, senderID))
	if _, err := f.client.SendFriendRequest(ctx, &pb.FriendRequest{UserId: senderID, ReceiverId: receiverID}); err != nil {
		logger.Error(ctx, "Failed to send friend request: %v", err)
		return err
	}

	return nil
}

func (f *FriendsClient) AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	logger.Info(ctx, fmt.Sprintf("Accepting friend request from %v to %v", senderID, receiverID))
	if _, err := f.client.AcceptFriendRequest(ctx, &pb.FriendRequest{UserId: senderID, ReceiverId: receiverID}); err != nil {
		logger.Error(ctx, "Failed to accept friend request: %v", err)
		return err
	}

	return nil
}

func (f *FriendsClient) Unfollow(ctx context.Context, userID string, friendID string) error {
	logger.Info(ctx, fmt.Sprintf("Unfollowing %v from %v", friendID, userID))
	if _, err := f.client.Unfollow(ctx, &pb.FriendRequest{UserId: userID, ReceiverId: friendID}); err != nil {
		logger.Error(ctx, "Failed to unfollow: %v", err)
		return err
	}

	return nil
}

func (f *FriendsClient) DeleteFriend(ctx context.Context, user string, friend string) error {
	logger.Error(ctx, fmt.Sprintf("Deleting friend %v from %v", friend, user))
	if _, err := f.client.DeleteFriend(ctx, &pb.FriendRequest{UserId: user, ReceiverId: friend}); err != nil {
		logger.Error(ctx, "Failed to delete friend: %v", err)
		return err
	}

	return nil
}
