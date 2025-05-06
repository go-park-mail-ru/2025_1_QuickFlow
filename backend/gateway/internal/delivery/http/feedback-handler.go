package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/forms"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type FeedbackUseCase interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
	GetNumMessagesSent(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumPostsCreated(ctx context.Context, userId uuid.UUID) (int64, error)
	GetNumProfileChanges(ctx context.Context, userId uuid.UUID) (int64, error)
}

type FeedbackHandler struct {
	feedbackUseCase FeedbackUseCase
	profileService  ProfileUseCase
	policy          *bluemonday.Policy
}

func NewFeedbackHandler(feedbackUseCase FeedbackUseCase, profileService ProfileUseCase, policy *bluemonday.Policy) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackUseCase: feedbackUseCase,
		profileService:  profileService,
		policy:          policy,
	}
}

func (f *FeedbackHandler) SaveFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while saving feedback")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	var form forms.FeedbackForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		logger.Error(ctx, "Failed to decode request body for feedback", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Bad request body", http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeFeedbackText(&form, f.policy)

	feedback, err := form.ToFeedback(user.Id)
	if err != nil {
		logger.Error(ctx, "Invalid feedback form", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid feedback form", http.StatusBadRequest))
		return
	}

	if err := f.feedbackUseCase.SaveFeedback(ctx, feedback); err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to save feedback", err)
		http2.WriteJSONError(w, appErr)
		return
	}
}

func (f *FeedbackHandler) GetAllFeedbackType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var chatForm forms.GetFeedbackForm
	if err := chatForm.GetParams(r.URL.Query()); err != nil {
		logger.Error(ctx, "Failed to parse query params for feedback list", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
		return
	}

	feedbacks, err := f.feedbackUseCase.GetAllFeedbackType(ctx, chatForm.Type, chatForm.Ts, chatForm.Count)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch feedback list", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	profileInfos := make(map[uuid.UUID]models.PublicUserInfo)
	var avg float64
	var feedbackOutput []forms.FeedbackFormOut

	for _, feedback := range feedbacks {
		info, found := profileInfos[feedback.RespondentId]
		if !found {
			info, err = f.profileService.GetPublicUserInfo(ctx, feedback.RespondentId)
			if err != nil {
				appErr := errors2.FromGRPCError(err)
				logger.Error(ctx, "Failed to load respondent info", err)
				http2.WriteJSONError(w, appErr)
				return
			}
			profileInfos[feedback.RespondentId] = info
		}
		feedbackOutput = append(feedbackOutput, forms.FromFeedBack(feedback, info))
		avg += float64(feedback.Rating)
	}

	if len(feedbackOutput) > 0 {
		avg /= float64(len(feedbackOutput))
	}

	result := forms.FeedbackOutAverage{Feedbacks: feedbackOutput, Average: avg}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.FeedbackOutAverage]{Payload: result}); err != nil {
		logger.Error(ctx, "Failed to encode feedback output", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
	}
}

func (f *FeedbackHandler) GetNumMessagesSent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching messages count")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	num, err := f.feedbackUseCase.GetNumMessagesSent(ctx, user.Id)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch number of messages sent", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}}); err != nil {
		logger.Error(ctx, "Failed to encode messages count", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode messages count", http.StatusInternalServerError))
	}
}

func (f *FeedbackHandler) GetNumPostsCreated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching posts count")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	num, err := f.feedbackUseCase.GetNumPostsCreated(ctx, user.Id)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch number of posts created", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}}); err != nil {
		logger.Error(ctx, "Failed to encode posts count", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode posts count", http.StatusInternalServerError))
	}
}

func (f *FeedbackHandler) GetNumProfileChanges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching profile changes")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	num, err := f.feedbackUseCase.GetNumProfileChanges(ctx, user.Id)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch number of profile changes", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}}); err != nil {
		logger.Error(ctx, "Failed to encode profile changes count", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode profile changes count", http.StatusInternalServerError))
	}
}
