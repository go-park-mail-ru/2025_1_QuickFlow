package friends_service

import (
	"github.com/google/uuid"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
)

func fromModelFriendInfoToGrpc(info models.FriendInfo) *pb.GetFriendInfo {
	return &pb.GetFriendInfo{
		Id:         info.Id.String(),
		Username:   info.Username,
		Firstname:  info.Firstname,
		Lastname:   info.Lastname,
		AvatarUrl:  info.AvatarURL,
		University: info.University,
	}
}

func fromGrpcToModelFriendInfo(info *pb.GetFriendInfo) models.FriendInfo {
	return models.FriendInfo{
		Id:         uuid.MustParse(info.Id),
		Firstname:  info.Firstname,
		Lastname:   info.Lastname,
		Username:   info.Username,
		AvatarURL:  info.AvatarUrl,
		University: info.University,
	}
}

func FromModelFriendsInfoToGrpc(infos []models.FriendInfo, friendsCount int) *pb.GetFriendsInfoResponse {
	var friendsInfos []*pb.GetFriendInfo
	for _, info := range infos {
		friendsInfos = append(friendsInfos, fromModelFriendInfoToGrpc(info))
	}

	return &pb.GetFriendsInfoResponse{
		Friends:    friendsInfos,
		TotalCount: int32(friendsCount),
	}
}

func FromGrpcToModelFriendsInfo(in *pb.GetFriendsInfoResponse) ([]models.FriendInfo, int) {
	result := make([]models.FriendInfo, len(in.Friends))
	for i, info := range in.Friends {
		result[i] = fromGrpcToModelFriendInfo(info)
	}

	return result, int(in.TotalCount)
}
