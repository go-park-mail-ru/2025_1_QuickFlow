package dto

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "quickflow/messenger_service/internal/delivery/grpc/proto"
	"quickflow/shared/models"
)

func MapChatToProto(chat models.Chat) *pb.Chat {
	res := &pb.Chat{
		Id:        chat.ID.String(),
		Name:      chat.Name,
		Type:      pb.ChatType(chat.Type),
		AvatarUrl: chat.AvatarURL,
		CreatedAt: timestamppb.New(chat.CreatedAt),
		UpdatedAt: timestamppb.New(chat.UpdatedAt),
	}

	if chat.LastReadByOther != nil {
		res.LastReadByOthers = timestamppb.New(*chat.LastReadByOther)
	}

	if chat.LastReadByMe != nil {
		res.LastReadByMe = timestamppb.New(*chat.LastReadByMe)
	}

	if chat.LastMessage.ID != uuid.Nil {
		res.LastMessage = MapMessageToProto(chat.LastMessage)
	}

	return res
}

func MapChatsToProto(chats []models.Chat) []*pb.Chat {
	res := make([]*pb.Chat, len(chats))
	for i, chat := range chats {
		res[i] = MapChatToProto(chat)
	}
	return res
}

//func MapProtoToChat(chat *pb.Chat) (*models.Chat, error) {
//    id, err := uuid.Parse(chat.Id)
//    if err != nil {
//        return nil, err
//    }
//
//    return &models.Chat{
//        ID:        id,
//        Name:      chat.Name,
//        Type:      models.ChatType(chat.Type),
//        AvatarURL: chat.AvatarUrl,
//        CreatedAt: chat.CreatedAt.AsTime(),
//        UpdatedAt: chat.UpdatedAt.AsTime(),
//    }, nil
//}

func MapProtoCreationInfoToModel(chatInfo *pb.ChatCreationInfo) models.ChatCreationInfo {
	return models.ChatCreationInfo{
		Name:   chatInfo.Name,
		Type:   models.ChatType(chatInfo.Type),
		Avatar: ProtoFileToModel(chatInfo.Avatar),
	}
}
