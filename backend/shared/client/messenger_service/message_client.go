package messenger_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type MessageServiceClient struct {
	client pb.MessageServiceClient
}

func NewMessageServiceClient(conn *grpc.ClientConn) *MessageServiceClient {
	return &MessageServiceClient{
		client: pb.NewMessageServiceClient(conn),
	}
}

func (c *MessageServiceClient) SendMessage(ctx context.Context, msg *models.Message, userId uuid.UUID) (*models.Message, error) {
	protoMsg := MapMessageToProto(*msg)
	logger.Info(ctx, "Sending message: %s", protoMsg.String())
	resp, err := c.client.SendMessage(ctx, &pb.SendMessageRequest{Message: protoMsg, UserAuthId: userId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to send message: %v", err)
		return nil, err
	}
	return MapProtoToMessage(resp.Message)
}

func (c *MessageServiceClient) GetMessagesForChat(ctx context.Context, chatID uuid.UUID, num int, updatedAt time.Time, userId uuid.UUID) ([]*models.Message, error) {
	logger.Info(ctx, "Getting messages for chatId: %s", chatID.String())
	resp, err := c.client.GetMessagesForChat(ctx, &pb.GetMessagesForChatRequest{
		ChatId:      chatID.String(),
		MessagesNum: int32(num),
		UpdatedAt:   timestamppb.New(updatedAt),
		UserAuthId:  userId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get messages for chat: %v", err)
		return nil, err
	}
	messages := make([]*models.Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		msg, err := MapProtoToMessage(m)
		if err != nil {
			logger.Error(ctx, "Failed to convert message to proto: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (c *MessageServiceClient) GetMessageById(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
	logger.Info(ctx, "Getting message: %s", msgID.String())
	resp, err := c.client.GetMessageById(ctx, &pb.GetMessageByIdRequest{MessageId: msgID.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get message: %v", err)
		return nil, err
	}
	return MapProtoToMessage(resp.Message)
}

func (c *MessageServiceClient) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	logger.Info(ctx, "Deleting message: %s", msgID.String())
	_, err := c.client.DeleteMessage(ctx, &pb.DeleteMessageRequest{MessageId: msgID.String()})
	return err
}

func (c *MessageServiceClient) UpdateLastReadTs(ctx context.Context, chatID, userID uuid.UUID, ts time.Time, userAuthId uuid.UUID) error {
	logger.Info(ctx, "Updating last read timestamp for chatId: %s", chatID.String())
	_, err := c.client.UpdateLastReadTs(ctx, &pb.UpdateLastReadTsRequest{
		ChatId:            chatID.String(),
		UserId:            userID.String(),
		LastReadTimestamp: timestamppb.New(ts),
		UserAuthId:        userAuthId.String(),
	})
	return err
}

func (c *MessageServiceClient) GetLastReadTs(ctx context.Context, chatID, userID uuid.UUID) (time.Time, error) {
	logger.Info(ctx, "Getting last read timestamp for chatId: %s", chatID.String())
	resp, err := c.client.GetLastReadTs(ctx, &pb.GetLastReadTsRequest{
		ChatId: chatID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get last read timestamp: %v", err)
		return time.Time{}, err
	}
	return resp.LastReadTs.AsTime(), nil
}

func (c *MessageServiceClient) GetNumUnreadMessages(ctx context.Context, chatID, userID uuid.UUID) (int, error) {
	logger.Info(ctx, "Getting number of unread messages for chatId: %s", chatID.String())
	resp, err := c.client.GetNumUnreadMessages(ctx, &pb.GetNumUnreadMessagesRequest{
		ChatId: chatID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get number of unread messages: %v", err)
		return 0, err
	}
	return int(resp.NumMessages), nil
}
