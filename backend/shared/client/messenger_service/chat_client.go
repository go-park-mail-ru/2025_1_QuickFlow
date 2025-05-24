package messenger_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type ChatServiceClient struct {
	client pb.ChatServiceClient
}

func NewChatServiceClient(conn *grpc.ClientConn) *ChatServiceClient {
	return &ChatServiceClient{
		client: pb.NewChatServiceClient(conn),
	}
}

func (c *ChatServiceClient) GetUserChats(ctx context.Context, userId uuid.UUID, limit int, updatedAt time.Time) ([]models.Chat, error) {
	req := &pb.GetUserChatsRequest{
		UserId:    userId.String(),
		ChatsNum:  int32(limit),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	logger.Info(ctx, "Getting user chats for userId: %s", userId.String())
	resp, err := c.client.GetUserChats(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to get user chats: %v", err)
		return nil, err
	}
	return MapProtoToChats(resp.Chats), nil
}

func (c *ChatServiceClient) CreateChat(ctx context.Context, userId uuid.UUID, info models.ChatCreationInfo) (*models.Chat, error) {
	req := &pb.CreateChatRequest{
		UserId: userId.String(),
		ChatInfo: &pb.ChatCreationInfo{
			Name:   info.Name,
			Type:   pb.ChatType(info.Type),
			Avatar: file_service.ModelFileToProto(info.Avatar),
		},
	}

	logger.Info(ctx, "Creating chat for userId: %s", userId.String())
	resp, err := c.client.CreateChat(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to create chat: %v", err)
		return nil, err
	}
	return MapProtoToChat(resp.Chat), nil
}

func (c *ChatServiceClient) GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]uuid.UUID, error) {
	logger.Info(ctx, "Getting chat participants for chatId: %s", chatId.String())
	resp, err := c.client.GetChatParticipants(ctx, &pb.GetChatParticipantsRequest{ChatId: chatId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get chat participants: %v", err)
		return nil, err
	}
	ids := make([]uuid.UUID, len(resp.ParticipantIds))
	for i, s := range resp.ParticipantIds {
		id, err := uuid.Parse(s)
		if err != nil {
			logger.Error(ctx, "Failed to parse participant ID: %v", err)
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

func (c *ChatServiceClient) GetPrivateChat(ctx context.Context, user1, user2 uuid.UUID) (*models.Chat, error) {
	logger.Info(ctx, "Getting private chat between %s and %s", user1.String(), user2.String())
	resp, err := c.client.GetPrivateChat(ctx, &pb.GetPrivateChatRequest{
		User1Id: user1.String(),
		User2Id: user2.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get private chat: %v", err)
		return nil, err
	}
	return MapProtoToChat(resp.Chat), nil
}

func (c *ChatServiceClient) DeleteChat(ctx context.Context, chatId uuid.UUID) error {
	logger.Info(ctx, "Deleting chat for userId: %s", chatId.String())
	_, err := c.client.DeleteChat(ctx, &pb.DeleteChatRequest{ChatId: chatId.String()})
	return err
}

func (c *ChatServiceClient) GetChat(ctx context.Context, chatId uuid.UUID) (*models.Chat, error) {
	logger.Info(ctx, "Getting chat for chatId: %s", chatId.String())
	resp, err := c.client.GetChat(ctx, &pb.GetChatRequest{ChatId: chatId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get chat: %v", err)
		return nil, err
	}
	return MapProtoToChat(resp.Chat), nil
}

func (c *ChatServiceClient) JoinChat(ctx context.Context, chatId, userId uuid.UUID) error {
	logger.Info(ctx, "Joining chat for userId: %s", userId.String())
	_, err := c.client.JoinChat(ctx, &pb.JoinChatRequest{
		ChatId: chatId.String(),
		UserId: userId.String(),
	})
	return err
}

func (c *ChatServiceClient) LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error {
	logger.Info(ctx, "Leaving chat for userId: %s", userId.String())
	_, err := c.client.LeaveChat(ctx, &pb.LeaveChatRequest{
		ChatId: chatId.String(),
		UserId: userId.String(),
	})
	return err
}

func (c *ChatServiceClient) GetNumUnreadChats(ctx context.Context, userId uuid.UUID) (int, error) {
	logger.Info(ctx, "Getting number of unread chats for userId: %s", userId.String())
	resp, err := c.client.GetNumUnreadChats(ctx, &pb.GetNumUnreadChatsRequest{UserId: userId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get number of unread chats: %v", err)
		return 0, err
	}
	return int(resp.NumChats), nil
}
