package messenger_service

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
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

func MapProtoToChat(chat *pb.Chat) *models.Chat {
	id, err := uuid.Parse(chat.Id)
	if err != nil {
		return nil
	}

	res := &models.Chat{
		ID:        id,
		Name:      chat.Name,
		Type:      models.ChatType(chat.Type),
		AvatarURL: chat.AvatarUrl,
		CreatedAt: chat.CreatedAt.AsTime(),
		UpdatedAt: chat.UpdatedAt.AsTime(),
	}

	if chat.LastReadByOthers != nil {
		tm := chat.LastReadByOthers.AsTime()
		res.LastReadByOther = &tm
	}

	if chat.LastReadByMe != nil {
		tm := chat.LastReadByMe.AsTime()
		res.LastReadByMe = &tm
	}

	if chat.LastMessage != nil {
		msg, err := MapProtoToMessage(chat.LastMessage)
		if err == nil {
			res.LastMessage = *msg
		}
	}

	return res
}

func MapProtoToChats(chats []*pb.Chat) []models.Chat {
	res := make([]models.Chat, len(chats))
	for i, chat := range chats {
		res[i] = *MapProtoToChat(chat)
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
		Avatar: file_service.ProtoFileToModel(chatInfo.Avatar),
	}
}
