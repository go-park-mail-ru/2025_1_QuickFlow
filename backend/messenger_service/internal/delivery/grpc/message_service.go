package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	dto "quickflow/shared/client/messenger_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type MessageUseCase interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (models.Message, error)
	GetMessagesForChatOlder(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, numMessages int, timestamp time.Time) ([]models.Message, error)
	SaveMessage(ctx context.Context, message models.Message) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageId uuid.UUID) error
	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*time.Time, error)
	UpdateLastReadTs(ctx context.Context, timestamp time.Time, chatId uuid.UUID, userId uuid.UUID) error
}

type MessageServiceServer struct {
	pb.UnimplementedMessageServiceServer
	MessageUseCase MessageUseCase
}

func NewMessageServiceServer(messageUseCase MessageUseCase) *MessageServiceServer {
	return &MessageServiceServer{MessageUseCase: messageUseCase}
}

func (m *MessageServiceServer) GetMessagesForChat(ctx context.Context, req *pb.GetMessagesForChatRequest) (*pb.GetMessagesForChatResponse, error) {
	logger.Info(ctx, "GetMessagesForChat request received")
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid chatId: ", err)
		return nil, err
	}

	userId, err := uuid.Parse(req.UserAuthId)
	if err != nil {
		logger.Error(ctx, "Invalid userAuthId: ", err)
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	messages, err := m.MessageUseCase.GetMessagesForChatOlder(ctx, chatId, userId, int(req.MessagesNum), req.UpdatedAt.AsTime())
	if err != nil {
		logger.Error(ctx, "Failed to get messages: ", err)
		return nil, err
	}

	return &pb.GetMessagesForChatResponse{Messages: dto.MapMessagesToProto(messages)}, nil
}

func (m *MessageServiceServer) GetMessageById(ctx context.Context, req *pb.GetMessageByIdRequest) (*pb.GetMessageByIdResponse, error) {
	logger.Info(ctx, "GetMessageById request received")
	messageId, err := uuid.Parse(req.MessageId)
	if err != nil {
		logger.Error(ctx, "Invalid messageId: ", err)
		return nil, err
	}

	message, err := m.MessageUseCase.GetMessageById(ctx, messageId)
	if err != nil {
		logger.Error(ctx, "Failed to get message by ID: ", err)
		return nil, err
	}

	return &pb.GetMessageByIdResponse{Message: dto.MapMessageToProto(message)}, nil
}

func (m *MessageServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	logger.Info(ctx, "SendMessage request received")
	userId, err := uuid.Parse(req.UserAuthId)
	if err != nil {
		logger.Error(ctx, "Invalid userAuthId: ", err)
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	message, err := dto.MapProtoToMessage(req.Message)
	if err != nil {
		logger.Error(ctx, "Failed to map proto to message: ", err)
		return nil, err
	}
	message.SenderID = userId

	savedMessage, err := m.MessageUseCase.SaveMessage(ctx, *message)
	if err != nil {
		logger.Error(ctx, "Failed to save message: ", err)
		return nil, err
	}

	return &pb.SendMessageResponse{Message: dto.MapMessageToProto(*savedMessage)}, nil
}

func (m *MessageServiceServer) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error) {
	logger.Info(ctx, "DeleteMessage request received")
	messageId, err := uuid.Parse(req.MessageId)
	if err != nil {
		logger.Error(ctx, "Invalid messageId: ", err)
		return nil, err
	}

	err = m.MessageUseCase.DeleteMessage(ctx, messageId)
	if err != nil {
		logger.Error(ctx, "Failed to delete message: ", err)
		return nil, err
	}

	return &pb.DeleteMessageResponse{Success: true}, nil
}

func (m *MessageServiceServer) UpdateLastReadTs(ctx context.Context, req *pb.UpdateLastReadTsRequest) (*pb.UpdateLastReadTsResponse, error) {
	logger.Info(ctx, "UpdateLastReadTs request received")
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid chatId: ", err)
		return nil, err
	}

	userId, err := uuid.Parse(req.UserAuthId)
	if err != nil {
		logger.Error(ctx, "Invalid userAuthId: ", err)
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	err = m.MessageUseCase.UpdateLastReadTs(ctx, req.LastReadTimestamp.AsTime(), chatId, userId)
	if err != nil {
		logger.Error(ctx, "Failed to update last read timestamp: ", err)
		return nil, err
	}

	return &pb.UpdateLastReadTsResponse{Success: true}, nil
}

func (m *MessageServiceServer) GetLastReadTs(ctx context.Context, req *pb.GetLastReadTsRequest) (*pb.GetLastReadTsResponse, error) {
	logger.Info(ctx, "GetLastReadTs request received")
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		logger.Error(ctx, "Invalid chatId: ", err)
		return nil, err
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid userId: ", err)
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	lastReadTs, err := m.MessageUseCase.GetLastReadTs(ctx, chatId, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get last read timestamp: ", err)
		return nil, err
	}

	return &pb.GetLastReadTsResponse{
		LastReadTs: timestamppb.New(*lastReadTs),
	}, nil
}
