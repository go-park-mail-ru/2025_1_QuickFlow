package friends_service

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"

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

func (f *FriendsClient) GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]shared_models.FriendInfo, int, error) {
	info, err := f.client.GetFriendsInfo(ctx, &pb.GetFriendsInfoRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, 0, err
	}

	res, friendsCount := FromGrpcToModelFriendsInfo(info)

	return res, friendsCount, nil
}

func (f *FriendsClient) GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (shared_models.UserRelation, error) {
	relation, err := f.client.GetUserRelation(ctx, &pb.FriendRequest{UserId: user1.String(), ReceiverId: user2.String()})
	if err != nil {
		return "", err
	}

	return shared_models.UserRelation(relation.Relation), nil
}

func (f *FriendsClient) SendFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	if _, err := f.client.SendFriendRequest(ctx, &pb.FriendRequest{UserId: senderID, ReceiverId: receiverID}); err != nil {
		return err
	}

	return nil
}

func (f *FriendsClient) AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	if _, err := f.client.AcceptFriendRequest(ctx, &pb.FriendRequest{UserId: senderID, ReceiverId: receiverID}); err != nil {
		return err
	}

	return nil
}

func (f *FriendsClient) Unfollow(ctx context.Context, userID string, friendID string) error {
	if _, err := f.client.Unfollow(ctx, &pb.FriendRequest{UserId: userID, ReceiverId: friendID}); err != nil {
		return err
	}

	return nil
}

func (f *FriendsClient) DeleteFriend(ctx context.Context, user string, friend string) error {
	if _, err := f.client.DeleteFriend(ctx, &pb.FriendRequest{UserId: user, ReceiverId: friend}); err != nil {
		return err
	}

	return nil
}
