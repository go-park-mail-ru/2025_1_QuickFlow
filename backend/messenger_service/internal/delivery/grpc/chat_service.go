package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/messenger_service/internal/delivery/grpc/dto"
	pb "quickflow/messenger_service/internal/delivery/grpc/proto"
	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/shared/models"
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
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	chats, err := c.chatUseCase.GetUserChats(ctx, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetUserChatsResponse{
		Chats: dto.MapChatsToProto(chats),
	}, nil
}

func (c *ChatServiceServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	chatInfo := dto.MapProtoCreationInfoToModel(req.ChatInfo)

	chat, err := c.chatUseCase.CreateChat(ctx, chatInfo)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.CreateChatResponse{
		Chat: dto.MapChatToProto(chat),
	}, nil
}

func (c *ChatServiceServer) GetChatParticipants(ctx context.Context, req *pb.GetChatParticipantsRequest) (*pb.GetChatParticipantsResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	participants, err := c.chatUseCase.GetChatParticipants(ctx, chatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	var ids []string
	for _, id := range participants {
		ids = append(ids, id.String())
	}
	res := &pb.GetChatParticipantsResponse{
		ParticipantIds: ids,
	}

	return res, nil
}

func (c *ChatServiceServer) GetPrivateChat(ctx context.Context, req *pb.GetPrivateChatRequest) (*pb.GetPrivateChatResponse, error) {
	userId1, err := uuid.Parse(req.User1Id)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId2, err := uuid.Parse(req.User2Id)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	chat, err := c.chatUseCase.GetPrivateChat(ctx, userId1, userId2)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetPrivateChatResponse{
		Chat: dto.MapChatToProto(chat),
	}, nil
}

func (c *ChatServiceServer) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.DeleteChatResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	err = c.chatUseCase.DeleteChat(ctx, chatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.DeleteChatResponse{Success: true}, nil
}

func (c *ChatServiceServer) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.GetChatResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	chat, err := c.chatUseCase.GetChat(ctx, chatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetChatResponse{
		Chat: dto.MapChatToProto(chat),
	}, nil
}

func (c *ChatServiceServer) JoinChat(ctx context.Context, req *pb.JoinChatRequest) (*pb.JoinChatResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	err = c.chatUseCase.JoinChat(ctx, chatId, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.JoinChatResponse{Success: true}, nil
}

func (c *ChatServiceServer) LeaveChat(ctx context.Context, req *pb.LeaveChatRequest) (*pb.LeaveChatResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	err = c.chatUseCase.LeaveChat(ctx, chatId, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.LeaveChatResponse{Success: true}, nil
}

func grpcErrorFromAppError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, messenger_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, messenger_errors.ErrNotParticipant):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, messenger_errors.ErrInvalidChatCreationInfo):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, messenger_errors.ErrAlreadyInChat):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, messenger_errors.ErrInvalidChatType):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, messenger_errors.ErrInvalidNumMessages):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
