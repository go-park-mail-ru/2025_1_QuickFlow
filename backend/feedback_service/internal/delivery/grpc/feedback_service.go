package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/feedback_service/internal/delivery/grpc/dto"
	feedback_errors "quickflow/feedback_service/internal/errors"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/feedback_service"
)

type FeedbackService interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
	GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error)
}

type ProfileService interface {
	GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models.PublicUserInfo, error)
}

type FeedBackService struct {
	pb.UnimplementedFeedbackServiceServer
	feedbackService FeedbackService
	profileService  ProfileService
}

func NewFeedbackServiceServer(feedbackService FeedbackService, profileService ProfileService) *FeedBackService {
	return &FeedBackService{
		feedbackService: feedbackService,
		profileService:  profileService,
	}
}

func (s *FeedBackService) SaveFeedback(ctx context.Context, req *pb.SaveFeedbackRequest) (*pb.SaveFeedbackResponse, error) {
	logger.Info(ctx, "Received SaveFeedback request")

	feedback, err := dto.ProtoFeedbackToModel(req.Feedback)
	if err != nil {
		logger.Error(ctx, "Failed to convert proto to model: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	err = s.feedbackService.SaveFeedback(ctx, feedback)
	if err != nil {
		logger.Error(ctx, "Failed to save feedback: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	logger.Info(ctx, "Successfully saved feedback")
	return &pb.SaveFeedbackResponse{Success: true}, nil
}

func (s *FeedBackService) GetAllFeedbackType(ctx context.Context, req *pb.GetAllFeedbackTypeRequest) (*pb.GetAllFeedbackTypeResponse, error) {
	logger.Info(ctx, "Received GetAllFeedbackType request")

	feedbackType, err := dto.FeedBackTypeFromProto(req.Type)
	if err != nil {
		logger.Error(ctx, "Invalid feedback type: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	feedbacks, err := s.feedbackService.GetAllFeedbackType(ctx, feedbackType, req.Ts.AsTime(), int(req.Count))
	if err != nil {
		logger.Error(ctx, "Failed to get feedbacks: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	protoFeedbacks := make([]*pb.Feedback, len(feedbacks))
	for i, feedback := range feedbacks {
		protoFeedbacks[i], err = dto.ModelFeedbackToProto(&feedback)
		if err != nil {
			logger.Error(ctx, "Failed to convert feedback model to proto: ", err)
			return nil, grpcErrorFromAppError(err)
		}
	}

	logger.Info(ctx, "Successfully fetched feedbacks")
	return &pb.GetAllFeedbackTypeResponse{Feedback: protoFeedbacks}, nil
}

func (s *FeedBackService) GetNumMessagesSent(ctx context.Context, req *pb.GetNumMessagesSentRequest) (*pb.GetNumMessagesSentResponse, error) {
	logger.Info(ctx, "Received GetNumMessagesSent request")

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	numMessages, err := s.feedbackService.GetNumMessagesSent(ctx, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get number of messages sent: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	logger.Info(ctx, "Successfully retrieved message count")
	return &pb.GetNumMessagesSentResponse{NumMessagesSent: numMessages}, nil
}

func (s *FeedBackService) GetNumPostsCreated(ctx context.Context, req *pb.GetNumPostsCreatedRequest) (*pb.GetNumPostsCreatedResponse, error) {
	logger.Info(ctx, "Received GetNumPostsCreated request")

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	numPosts, err := s.feedbackService.GetNumPostsCreated(ctx, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get number of posts created: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	logger.Info(ctx, "Successfully retrieved post count")
	return &pb.GetNumPostsCreatedResponse{NumPostsCreated: numPosts}, nil
}

func (s *FeedBackService) GetNumProfileChanges(ctx context.Context, req *pb.GetNumProfileChangesRequest) (*pb.GetNumProfileChangesResponse, error) {
	logger.Info(ctx, "Received GetNumProfileChanges request")

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	numProfileChanges, err := s.feedbackService.GetNumProfileChanges(ctx, userId)
	if err != nil {
		logger.Error(ctx, "Failed to get number of profile changes: ", err)
		return nil, grpcErrorFromAppError(err)
	}

	logger.Info(ctx, "Successfully retrieved profile change count")
	return &pb.GetNumProfileChangesResponse{NumProfileChanges: numProfileChanges}, nil
}

func grpcErrorFromAppError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, feedback_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, feedback_errors.ErrRespondent) ||
		errors.Is(err, feedback_errors.ErrRating) ||
		errors.Is(err, feedback_errors.ErrTextTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
