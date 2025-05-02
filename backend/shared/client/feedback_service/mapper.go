package feedback_service

import (
	"errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/feedback_service"
)

func ProtoFeedbackToModel(f *pb.Feedback) (*models.Feedback, error) {
	id, err := uuid.Parse(f.Id)
	if err != nil {
		return nil, err
	}
	respondentId, err := uuid.Parse(f.RespondentId)
	if err != nil {
		return nil, err
	}

	feedbackType, err := FeedBackTypeFromProto(f.Type)
	if err != nil {
		return nil, err
	}
	return &models.Feedback{
		Id:           id,
		Rating:       int(f.Rating),
		Text:         f.Text,
		Type:         feedbackType,
		RespondentId: respondentId,
		CreatedAt:    f.CreatedAt.AsTime(),
	}, nil
}

func FeedBackTypeToProto(f models.FeedbackType) (pb.FeedbackType, error) {
	switch f {
	case models.FeedbackGeneral:
		return pb.FeedbackType_FEEDBACK_GENERAL, nil
	case models.FeedbackPost:
		return pb.FeedbackType_FEEDBACK_POST, nil
	case models.FeedbackMessenger:
		return pb.FeedbackType_FEEDBACK_MESSENGER, nil
	case models.FeedbackRecommendation:
		return pb.FeedbackType_FEEDBACK_RECOMMENDATIONS, nil
	case models.FeedbackProfile:
		return pb.FeedbackType_FEEDBACK_PROFILE, nil
	case models.FeedbackAuth:
		return pb.FeedbackType_FEEDBACK_AUTH, nil
	default:
		return pb.FeedbackType_FEEDBACK_GENERAL, errors.New("invalid feedback type")
	}
}

func FeedBackTypeFromProto(f pb.FeedbackType) (models.FeedbackType, error) {
	switch f {
	case pb.FeedbackType_FEEDBACK_GENERAL:
		return models.FeedbackGeneral, nil
	case pb.FeedbackType_FEEDBACK_POST:
		return models.FeedbackPost, nil
	case pb.FeedbackType_FEEDBACK_MESSENGER:
		return models.FeedbackMessenger, nil
	case pb.FeedbackType_FEEDBACK_RECOMMENDATIONS:
		return models.FeedbackRecommendation, nil
	case pb.FeedbackType_FEEDBACK_PROFILE:
		return models.FeedbackProfile, nil
	case pb.FeedbackType_FEEDBACK_AUTH:
		return models.FeedbackAuth, nil
	default:
		return "", errors.New("invalid feedback type")
	}
}

func ModelFeedbackToProto(f *models.Feedback) (*pb.Feedback, error) {
	feedbackType, err := FeedBackTypeToProto(f.Type)
	if err != nil {
		return nil, err
	}

	return &pb.Feedback{
		Id:           f.Id.String(),
		Rating:       int32(f.Rating),
		Text:         f.Text,
		Type:         feedbackType,
		RespondentId: f.RespondentId.String(),
		CreatedAt:    timestamppb.New(f.CreatedAt),
	}, nil
}
