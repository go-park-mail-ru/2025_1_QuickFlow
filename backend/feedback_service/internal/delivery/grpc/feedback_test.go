package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/feedback_service/internal/delivery/grpc/mocks"
	feedback_errors "quickflow/feedback_service/internal/errors"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/feedback_service"
)

func TestSaveFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now().UTC()
	feedbackID := uuid.New()
	respondentID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.SaveFeedbackRequest
		mockSetup   func(*mocks.MockFeedbackService)
		want        *pb.SaveFeedbackResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful feedback save",
			req: &pb.SaveFeedbackRequest{
				Feedback: &pb.Feedback{
					Id:           feedbackID.String(),
					Rating:       5,
					RespondentId: respondentID.String(),
					Text:         "Great service!",
					Type:         pb.FeedbackType_FEEDBACK_GENERAL,
					CreatedAt:    timestamppb.New(now),
				},
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().SaveFeedback(gomock.Any(), &models.Feedback{
					Id:           feedbackID,
					Rating:       5,
					RespondentId: respondentID,
					Text:         "Great service!",
					Type:         models.FeedbackGeneral,
					CreatedAt:    now,
				}).Return(nil)
			},
			want:    &pb.SaveFeedbackResponse{Success: true},
			wantErr: false,
		},
		{
			name: "invalid uuid format",
			req: &pb.SaveFeedbackRequest{
				Feedback: &pb.Feedback{
					Id: "invalid-uuid",
				},
			},
			mockSetup:   func(m *mocks.MockFeedbackService) {},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "invalid UUID length: 12"),
		},
		{
			name: "service returns error",
			req: &pb.SaveFeedbackRequest{
				Feedback: &pb.Feedback{
					Id:           feedbackID.String(),
					Rating:       5,
					RespondentId: respondentID.String(),
					Text:         "Great service!",
					Type:         pb.FeedbackType_FEEDBACK_GENERAL,
					CreatedAt:    timestamppb.New(now),
				},
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().SaveFeedback(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeedbackService := mocks.NewMockFeedbackService(ctrl)
			mockProfileService := mocks.NewMockProfileService(ctrl)
			tt.mockSetup(mockFeedbackService)

			server := NewFeedbackServiceServer(mockFeedbackService, mockProfileService)
			resp, err := server.SaveFeedback(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGetAllFeedbackType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now().UTC()

	tests := []struct {
		name        string
		req         *pb.GetAllFeedbackTypeRequest
		mockSetup   func(*mocks.MockFeedbackService)
		want        *pb.GetAllFeedbackTypeResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "service returns error",
			req: &pb.GetAllFeedbackTypeRequest{
				Type:  pb.FeedbackType_FEEDBACK_GENERAL,
				Ts:    timestamppb.New(now),
				Count: 10,
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetAllFeedbackType(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return([]models.Feedback{}, errors.New("database error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeedbackService := mocks.NewMockFeedbackService(ctrl)
			mockProfileService := mocks.NewMockProfileService(ctrl)
			tt.mockSetup(mockFeedbackService)

			server := NewFeedbackServiceServer(mockFeedbackService, mockProfileService)
			resp, err := server.GetAllFeedbackType(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGetNumMessagesSent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.GetNumMessagesSentRequest
		mockSetup   func(*mocks.MockFeedbackService)
		want        *pb.GetNumMessagesSentResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful get message count",
			req: &pb.GetNumMessagesSentRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumMessagesSent(gomock.Any(), userID).Return(int64(42), nil)
			},
			want: &pb.GetNumMessagesSentResponse{
				NumMessagesSent: 42,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetNumMessagesSentRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func(m *mocks.MockFeedbackService) {},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "invalid UUID length: 12"),
		},
		{
			name: "service returns error",
			req: &pb.GetNumMessagesSentRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumMessagesSent(gomock.Any(), userID).Return(int64(0), errors.New("database error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "database error"),
		},
		{
			name: "user not found",
			req: &pb.GetNumMessagesSentRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumMessagesSent(gomock.Any(), userID).Return(int64(0), feedback_errors.ErrNotFound)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.NotFound, feedback_errors.ErrNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeedbackService := mocks.NewMockFeedbackService(ctrl)
			mockProfileService := mocks.NewMockProfileService(ctrl)
			tt.mockSetup(mockFeedbackService)

			server := NewFeedbackServiceServer(mockFeedbackService, mockProfileService)
			resp, err := server.GetNumMessagesSent(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGetNumPostsCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.GetNumPostsCreatedRequest
		mockSetup   func(*mocks.MockFeedbackService)
		want        *pb.GetNumPostsCreatedResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful get posts count",
			req: &pb.GetNumPostsCreatedRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumPostsCreated(gomock.Any(), userID).Return(int64(15), nil)
			},
			want: &pb.GetNumPostsCreatedResponse{
				NumPostsCreated: 15,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetNumPostsCreatedRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func(m *mocks.MockFeedbackService) {},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "invalid UUID length: 12"),
		},
		{
			name: "service returns error",
			req: &pb.GetNumPostsCreatedRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumPostsCreated(gomock.Any(), userID).Return(int64(0), errors.New("database error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeedbackService := mocks.NewMockFeedbackService(ctrl)
			mockProfileService := mocks.NewMockProfileService(ctrl)
			tt.mockSetup(mockFeedbackService)

			server := NewFeedbackServiceServer(mockFeedbackService, mockProfileService)
			resp, err := server.GetNumPostsCreated(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGetNumProfileChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.GetNumProfileChangesRequest
		mockSetup   func(*mocks.MockFeedbackService)
		want        *pb.GetNumProfileChangesResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful get profile changes count",
			req: &pb.GetNumProfileChangesRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumProfileChanges(gomock.Any(), userID).Return(int64(7), nil)
			},
			want: &pb.GetNumProfileChangesResponse{
				NumProfileChanges: 7,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetNumProfileChangesRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func(m *mocks.MockFeedbackService) {},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "invalid UUID length: 12"),
		},
		{
			name: "service returns error",
			req: &pb.GetNumProfileChangesRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumProfileChanges(gomock.Any(), userID).Return(int64(0), errors.New("database error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "database error"),
		},
		{
			name: "user not found",
			req: &pb.GetNumProfileChangesRequest{
				UserId: userID.String(),
			},
			mockSetup: func(m *mocks.MockFeedbackService) {
				m.EXPECT().GetNumProfileChanges(gomock.Any(), userID).Return(int64(0), feedback_errors.ErrNotFound)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: status.Error(codes.NotFound, feedback_errors.ErrNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeedbackService := mocks.NewMockFeedbackService(ctrl)
			mockProfileService := mocks.NewMockProfileService(ctrl)
			tt.mockSetup(mockFeedbackService)

			server := NewFeedbackServiceServer(mockFeedbackService, mockProfileService)
			resp, err := server.GetNumProfileChanges(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGrpcErrorFromAppError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected error
	}{
		{
			name:     "not found error",
			input:    feedback_errors.ErrNotFound,
			expected: status.Error(codes.NotFound, feedback_errors.ErrNotFound.Error()),
		},
		{
			name:     "respondent error",
			input:    feedback_errors.ErrRespondent,
			expected: status.Error(codes.InvalidArgument, feedback_errors.ErrRespondent.Error()),
		},
		{
			name:     "rating error",
			input:    feedback_errors.ErrRating,
			expected: status.Error(codes.InvalidArgument, feedback_errors.ErrRating.Error()),
		},
		{
			name:     "text too long error",
			input:    feedback_errors.ErrTextTooLong,
			expected: status.Error(codes.InvalidArgument, feedback_errors.ErrTextTooLong.Error()),
		},
		{
			name:     "generic error",
			input:    errors.New("some error"),
			expected: status.Error(codes.Internal, "some error"),
		},
		{
			name:     "nil error",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grpcErrorFromAppError(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Error(), result.Error())
				assert.Equal(t, status.Code(tt.expected), status.Code(result))
			}
		})
	}
}
