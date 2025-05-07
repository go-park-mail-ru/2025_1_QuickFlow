package grpc

import (
	"context"
	"quickflow/messenger_service/internal/delivery/grpc/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	dto "quickflow/shared/client/messenger_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func TestMessageServiceServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockMessageUseCase(ctrl)
	server := NewMessageServiceServer(mockUseCase)

	ctx := context.Background()
	now := time.Now()
	testMessage := models.Message{
		ID:        uuid.New(),
		Text:      "Test message",
		CreatedAt: now,
		UpdatedAt: now,
		SenderID:  uuid.New(),
		ChatID:    uuid.New(),
	}
	testProtoMessage := dto.MapMessageToProto(testMessage)

	tests := []struct {
		name        string
		mockSetup   func()
		req         interface{}
		wantResp    interface{}
		wantErr     bool
		expectedErr error
	}{
		{
			name: "GetMessagesForChat - Invalid ChatID",
			req: &pb.GetMessagesForChatRequest{
				ChatId:      "invalid",
				MessagesNum: 10,
				UpdatedAt:   timestamppb.New(now),
				UserAuthId:  testMessage.SenderID.String(),
			},
			wantErr: true,
		},
		{
			name: "GetMessagesForChat - Invalid UserAuthID",
			req: &pb.GetMessagesForChatRequest{
				ChatId:      testMessage.ChatID.String(),
				MessagesNum: 10,
				UpdatedAt:   timestamppb.New(now),
				UserAuthId:  "invalid",
			},
			wantErr:     true,
			expectedErr: status.Error(codes.Unauthenticated, "user not found in context"),
		},
		{
			name: "SendMessage - Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					SaveMessage(ctx, gomock.Any()).
					Return(&testMessage, nil)
			},
			req: &pb.SendMessageRequest{
				Message:    testProtoMessage,
				UserAuthId: testMessage.SenderID.String(),
			},
			wantResp: &pb.SendMessageResponse{
				Message: testProtoMessage,
			},
		},
		{
			name: "SendMessage - Invalid UserAuthID",
			req: &pb.SendMessageRequest{
				Message:    testProtoMessage,
				UserAuthId: "invalid",
			},
			wantErr:     true,
			expectedErr: status.Error(codes.Unauthenticated, "user not found in context"),
		},
		{
			name: "GetMessageById - Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetMessageById(ctx, testMessage.ID).
					Return(testMessage, nil)
			},
			req: &pb.GetMessageByIdRequest{
				MessageId: testMessage.ID.String(),
			},
			wantResp: &pb.GetMessageByIdResponse{
				Message: testProtoMessage,
			},
		},
		{
			name: "GetMessageById - Invalid MessageID",
			req: &pb.GetMessageByIdRequest{
				MessageId: "invalid",
			},
			wantErr: true,
		},
		{
			name: "DeleteMessage - Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					DeleteMessage(ctx, testMessage.ID).
					Return(nil)
			},
			req: &pb.DeleteMessageRequest{
				MessageId: testMessage.ID.String(),
			},
			wantResp: &pb.DeleteMessageResponse{
				Success: true,
			},
		},
		{
			name: "DeleteMessage - Invalid MessageID",
			req: &pb.DeleteMessageRequest{
				MessageId: "invalid",
			},
			wantErr: true,
		},
		{
			name: "UpdateLastReadTs - Invalid ChatID",
			req: &pb.UpdateLastReadTsRequest{
				ChatId:            "invalid",
				UserId:            testMessage.SenderID.String(),
				LastReadTimestamp: timestamppb.New(now),
				UserAuthId:        testMessage.SenderID.String(),
			},
			wantErr: true,
		},
		{
			name: "GetLastReadTs - Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetLastReadTs(ctx, testMessage.ChatID, testMessage.SenderID).
					Return(&now, nil)
			},
			req: &pb.GetLastReadTsRequest{
				ChatId: testMessage.ChatID.String(),
				UserId: testMessage.SenderID.String(),
			},
			wantResp: &pb.GetLastReadTsResponse{
				LastReadTs: timestamppb.New(now),
			},
		},
		{
			name: "GetLastReadTs - Invalid ChatID",
			req: &pb.GetLastReadTsRequest{
				ChatId: "invalid",
				UserId: testMessage.SenderID.String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			var resp interface{}
			var err error

			switch req := tt.req.(type) {
			case *pb.GetMessagesForChatRequest:
				resp, err = server.GetMessagesForChat(ctx, req)
			case *pb.SendMessageRequest:
				resp, err = server.SendMessage(ctx, req)
			case *pb.GetMessageByIdRequest:
				resp, err = server.GetMessageById(ctx, req)
			case *pb.DeleteMessageRequest:
				resp, err = server.DeleteMessage(ctx, req)
			case *pb.UpdateLastReadTsRequest:
				resp, err = server.UpdateLastReadTs(ctx, req)
			case *pb.GetLastReadTsRequest:
				resp, err = server.GetLastReadTs(ctx, req)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}
