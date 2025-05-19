package messenger_service

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func MapMessageToProto(message models.Message) *pb.Message {
	return &pb.Message{
		Id:          message.ID.String(),
		SenderId:    message.SenderID.String(),
		ChatId:      message.ChatID.String(),
		Text:        message.Text,
		CreatedAt:   timestamppb.New(message.CreatedAt),
		UpdatedAt:   timestamppb.New(message.UpdatedAt),
		Attachments: file_service.ModelFilesToProto(message.Attachments),
		ReceiverId:  message.ReceiverID.String(),
	}
}

func MapMessagesToProto(messages []models.Message) []*pb.Message {
	res := make([]*pb.Message, len(messages))
	for i, message := range messages {
		res[i] = MapMessageToProto(message)
	}
	return res
}

func MapProtoToMessage(message *pb.Message) (*models.Message, error) {
	id, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, err
	}

	senderId, err := uuid.Parse(message.SenderId)
	if err != nil {
		return nil, err
	}

	chatId, err1 := uuid.Parse(message.ChatId)
	if err1 != nil {
		chatId = uuid.Nil
	}
	receiverId, err2 := uuid.Parse(message.ReceiverId)
	if err1 != nil && err2 != nil {
		return nil, err1
	} else if err2 != nil {
		receiverId = uuid.Nil
	}

	return &models.Message{
		ID:          id,
		SenderID:    senderId,
		ChatID:      chatId,
		ReceiverID:  receiverId,
		Text:        message.Text,
		CreatedAt:   message.CreatedAt.AsTime(),
		UpdatedAt:   message.UpdatedAt.AsTime(),
		Attachments: file_service.ProtoFilesToModels(message.Attachments),
	}, nil
}
