package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/messenger_service/internal/delivery/grpc/dto"
	pb "quickflow/messenger_service/internal/delivery/grpc/proto"
	"quickflow/shared/models"
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
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	messages, err := m.MessageUseCase.GetMessagesForChatOlder(ctx, chatId, userId, int(req.MessagesNum), req.UpdatedAt.AsTime())
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetMessagesForChatResponse{
		Messages: dto.MapMessagesToProto(messages),
	}, nil
}

func (m *MessageServiceServer) GetMessageById(ctx context.Context, req *pb.GetMessageByIdRequest) (*pb.GetMessageByIdResponse, error) {
	messageId, err := uuid.Parse(req.MessageId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	message, err := m.MessageUseCase.GetMessageById(ctx, messageId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetMessageByIdResponse{
		Message: dto.MapMessageToProto(message),
	}, nil
}

func (m *MessageServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	userId, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	message, err := dto.MapProtoToMessage(req.Message)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	message.SenderID = userId

	savedMessage, err := m.MessageUseCase.SaveMessage(ctx, *message)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.SendMessageResponse{
		Message: dto.MapMessageToProto(*savedMessage),
	}, nil
}

func (m *MessageServiceServer) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error) {
	messageId, err := uuid.Parse(req.MessageId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	err = m.MessageUseCase.DeleteMessage(ctx, messageId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.DeleteMessageResponse{
		Success: true,
	}, nil
}

func (m *MessageServiceServer) UpdateLastReadTs(ctx context.Context, req *pb.UpdateLastReadTsRequest) (*pb.UpdateLastReadTsResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	err = m.MessageUseCase.UpdateLastReadTs(ctx, req.LastReadTimestamp.AsTime(), chatId, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.UpdateLastReadTsResponse{
		Success: true,
	}, nil
}

func (m *MessageServiceServer) GetLastReadTs(ctx context.Context, req *pb.GetLastReadTsRequest) (*pb.GetLastReadTsResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	userId, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	lastReadTs, err := m.MessageUseCase.GetLastReadTs(ctx, chatId, userId)
	if err != nil {
		return nil, grpcErrorFromAppError(err)
	}

	return &pb.GetLastReadTsResponse{
		LastReadTs: timestamppb.New(*lastReadTs),
	}, nil
}
