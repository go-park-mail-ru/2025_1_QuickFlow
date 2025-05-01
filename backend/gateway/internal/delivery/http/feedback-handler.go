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
	"quickflow/shared/models"
)

type FeedbackUseCase interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedback(ctx context.Context, ts time.Time, count int) (map[models.FeedbackType][]models.Feedback, error)
	GetAverageRatings(ctx context.Context) (map[models.FeedbackType]float64, error)
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

func NewFeedbackService(feedbackUseCase FeedbackUseCase, profileService ProfileUseCase, policy *bluemonday.Policy) *FeedbackHandler {
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
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var form forms.FeedbackForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
		return
	}

	sanitizer.SanitizeFeedbackText(&form, f.policy)

	feedback, err := form.ToFeedback(user.Id)
	if err != nil {
		http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
		return
	}

	err = f.feedbackUseCase.SaveFeedback(ctx, feedback)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
		return
	}
}

func (f *FeedbackHandler) GetAllFeedbackType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var chatForm forms.GetFeedbackForm
	err := chatForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	feedbacks, err := f.feedbackUseCase.GetAllFeedbackType(ctx, chatForm.Type, chatForm.Ts, chatForm.Count)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
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
				err := errors2.FromGRPCError(err)
				http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
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
		http2.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (f *FeedbackHandler) GetNumMessagesSent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	num, err := f.feedbackUseCase.GetNumMessagesSent(ctx, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}})
}

func (f *FeedbackHandler) GetNumPostsCreated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	num, err := f.feedbackUseCase.GetNumPostsCreated(ctx, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}})
}

func (f *FeedbackHandler) GetNumProfileChanges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	num, err := f.feedbackUseCase.GetNumProfileChanges(ctx, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, err.Error(), err.HTTPStatus)
		return
	}

	type out struct {
		Count int64 `json:"count"`
	}
	json.NewEncoder(w).Encode(forms.PayloadWrapper[out]{Payload: out{Count: num}})
}
