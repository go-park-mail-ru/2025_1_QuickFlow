package grpc

import (
	"context"

	"github.com/google/uuid"

	dto "quickflow/shared/client/messenger_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type ChatUseCase interface {
	CreateChat(ctx context.Context, chatInfo models.ChatCreationInfo) (models.Chat, error)
	GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]uuid.UUID, error)
	GetUserChats(ctx context.Context, userId uuid.UUID) ([]models.Chat, error)
	GetPrivateChat(ctx context.Context, userId1, userId2 uuid.UUID) (models.Chat, error)
	DeleteChat(ctx context.Context, chatId uuid.UUID) error
	GetChat(ctx context.Context, chatId uuid.UUID) (models.Chat, error)
	JoinChat(ctx context.Context, chatId, userId uuid.UUID) error
	LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error
}

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
	chatUseCase ChatUseCase
}

func NewChatServiceServer(chatUseCase ChatUseCase) *ChatServiceServer {
	return &ChatServiceServer{chatUseCase: chatUseCase}
}

func (c *ChatServiceServer) GetUserChats(ctx context.Context, req *pb.GetUserChatsRequest) (*pb.GetUserChatsResponse, error) {
	logger.Info(ctx, "Received GetUserChats request")

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: ", err)
		return nil, err
	}

	chats, err := c.chatUseCase.GetUserChats(ctx, userId)
	if err != nil {
		logger.Error(ctx, "GetUserChats failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully fetched user chats")
	return &pb.GetUserChatsResponse{Chats: dto.MapChatsToProto(chats)}, nil
}

func (c *ChatServiceServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	logger.Info(ctx, "Received CreateChat request")

	chatInfo := dto.MapProtoCreationInfoToModel(req.ChatInfo)

	chat, err := c.chatUseCase.CreateChat(ctx, chatInfo)
	if err != nil {
		logger.Error(ctx, "CreateChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully created chat")
	return &pb.CreateChatResponse{Chat: dto.MapChatToProto(chat)}, nil
}

func (c *ChatServiceServer) GetChatParticipants(ctx context.Context, req *pb.GetChatParticipantsRequest) (*pb.GetChatParticipantsResponse, error) {
	logger.Info(ctx, "Received GetChatParticipants request")

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid ChatId: ", err)
		return nil, err
	}

	participants, err := c.chatUseCase.GetChatParticipants(ctx, chatId)
	if err != nil {
		logger.Error(ctx, "GetChatParticipants failed: ", err)
		return nil, err
	}

	var ids []string
	for _, id := range participants {
		ids = append(ids, id.String())
	}

	logger.Info(ctx, "Successfully fetched chat participants")
	return &pb.GetChatParticipantsResponse{ParticipantIds: ids}, nil
}

func (c *ChatServiceServer) GetPrivateChat(ctx context.Context, req *pb.GetPrivateChatRequest) (*pb.GetPrivateChatResponse, error) {
	logger.Info(ctx, "Received GetPrivateChat request")

	userId1, err := uuid.Parse(req.User1Id)
	if err != nil {
		logger.Error(ctx, "Invalid User1Id: ", err)
		return nil, err
	}

	userId2, err := uuid.Parse(req.User2Id)
	if err != nil {
		logger.Error(ctx, "Invalid User2Id: ", err)
		return nil, err
	}

	chat, err := c.chatUseCase.GetPrivateChat(ctx, userId1, userId2)
	if err != nil {
		logger.Error(ctx, "GetPrivateChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully fetched private chat")
	return &pb.GetPrivateChatResponse{Chat: dto.MapChatToProto(chat)}, nil
}

func (c *ChatServiceServer) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.DeleteChatResponse, error) {
	logger.Info(ctx, "Received DeleteChat request")

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid ChatId: ", err)
		return nil, err
	}

	err = c.chatUseCase.DeleteChat(ctx, chatId)
	if err != nil {
		logger.Error(ctx, "DeleteChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully deleted chat")
	return &pb.DeleteChatResponse{Success: true}, nil
}

func (c *ChatServiceServer) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.GetChatResponse, error) {
	logger.Info(ctx, "Received GetChat request")

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid ChatId: ", err)
		return nil, err
	}

	chat, err := c.chatUseCase.GetChat(ctx, chatId)
	if err != nil {
		logger.Error(ctx, "GetChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully fetched chat")
	return &pb.GetChatResponse{Chat: dto.MapChatToProto(chat)}, nil
}

func (c *ChatServiceServer) JoinChat(ctx context.Context, req *pb.JoinChatRequest) (*pb.JoinChatResponse, error) {
	logger.Info(ctx, "Received JoinChat request")

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid ChatId: ", err)
		return nil, err
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: ", err)
		return nil, err
	}

	err = c.chatUseCase.JoinChat(ctx, chatId, userId)
	if err != nil {
		logger.Error(ctx, "JoinChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully joined chat")
	return &pb.JoinChatResponse{Success: true}, nil
}

func (c *ChatServiceServer) LeaveChat(ctx context.Context, req *pb.LeaveChatRequest) (*pb.LeaveChatResponse, error) {
	logger.Info(ctx, "Received LeaveChat request")

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid ChatId: ", err)
		return nil, err
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: ", err)
		return nil, err
	}

	err = c.chatUseCase.LeaveChat(ctx, chatId, userId)
	if err != nil {
		logger.Error(ctx, "LeaveChat failed: ", err)
		return nil, err
	}

	logger.Info(ctx, "Successfully left chat")
	return &pb.LeaveChatResponse{Success: true}, nil
}
