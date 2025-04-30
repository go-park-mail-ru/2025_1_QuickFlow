package dto

import (
	pb "quickflow/friends_service/internal/delivery/grpc/proto"
	"quickflow/shared/models"
)

func fromModelFriendInfoToGrpc(info models.FriendInfo) *pb.GetFriendInfo {
	return &pb.GetFriendInfo{
		Id: info.Id.String(),
		Username: info.Username,
		Firstname: info.Firstname,
		Lastname: info.Lastname,
		AvatarUrl: info.AvatarURL,
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
