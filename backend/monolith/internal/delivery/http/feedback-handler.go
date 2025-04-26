package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/internal/usecase"
	"quickflow/pkg/logger"
	http2 "quickflow/utils/http"
	"time"
)

type FeedbackUseCase interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedback(ctx context.Context, ts time.Time, count int) (map[models.FeedbackType][]models.Feedback, error)
	GetAverageRatings(ctx context.Context) (map[models.FeedbackType]float64, error)
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
}

type FeedbackHandler struct {
	feedbackUseCase FeedbackUseCase
	profileService  ProfileUseCase
}

func NewFeedbackService(feedbackUseCase FeedbackUseCase, profileService ProfileUseCase) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackUseCase: feedbackUseCase,
		profileService:  profileService,
	}
}

func (f *FeedbackHandler) SaveFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching messages")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var form forms.FeedbackForm

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		logger.Error(ctx, fmt.Sprintf("Decode error: %s", err.Error()))
		http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
		return
	}

	feedback, err := form.ToFeedback(user.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("ToFeedback error: %s", err.Error()))
		http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
	}

	err = f.feedbackUseCase.SaveFeedback(ctx, feedback)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("SaveFeedback error: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to save feedback", http.StatusInternalServerError)
	}

	logger.Info(ctx, fmt.Sprintf("Saved feedback: %s", feedback))
}

func (f *FeedbackHandler) GetAllFeedbackType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Info(ctx, "Got GetUserChats request")

	var chatForm forms.GetFeedbackForm
	err := chatForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse query params: %v", err))
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, "Got GetFeedbackTypes request")
	feedbacks, err := f.feedbackUseCase.GetAllFeedbackType(ctx, chatForm.Type, chatForm.Ts, chatForm.Count)
	if errors.Is(err, usecase.ErrNotFound) {
		logger.Info(ctx, "GetFeedbackType not found")
		http2.WriteJSONError(w, "GetFeedbackType not found", http.StatusNotFound)
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("GetFeedbackType error: %s", err.Error()))
		http2.WriteJSONError(w, "GetFeedbackType error", http.StatusInternalServerError)
	}

	// profile infos
	profileInfos := make(map[uuid.UUID]models.PublicUserInfo)

	var avg float64 = 0
	var feedbackOutput []forms.FeedbackFormOut
	for _, feedback := range feedbacks {

		if info, found := profileInfos[feedback.RespondentId]; found {
			feedbackOutput = append(feedbackOutput, forms.FromFeedBack(feedback, info))
		} else {
			publicInfo, err := f.profileService.GetPublicUserInfo(ctx, feedback.RespondentId)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("GetFeedbackType error: %s", err.Error()))
				http2.WriteJSONError(w, "GetFeedbackType error", http.StatusInternalServerError)
			}
			profileInfos[feedback.RespondentId] = publicInfo
			feedbackOutput = append(feedbackOutput, forms.FromFeedBack(feedback, publicInfo))
		}
		avg += float64(feedback.Rating)
	}

	if len(feedbackOutput) > 0 {
		avg /= float64(len(feedbackOutput))
	}
	feedBackOutAVG := forms.FeedbackOutAverage{
		Feedbacks: feedbackOutput,
		Average:   avg,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.FeedbackOutAverage]{Payload: feedBackOutAVG})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode chats: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode chats", http.StatusInternalServerError)
		return
	}
}
