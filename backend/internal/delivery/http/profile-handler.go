package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"quickflow/internal/usecase"
	"quickflow/pkg/logger"

	"github.com/gorilla/mux"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	http2 "quickflow/utils/http"
)

type ProfileUseCase interface {
	GetUserInfoByUserName(ctx context.Context, username string) (models.Profile, error)
	UpdateProfile(ctx context.Context, newProfile models.Profile) error
	GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models.PublicUserInfo, error)
	GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]models.PublicUserInfo, error)
	UpdateLastSeen(ctx context.Context, userId uuid.UUID) error
}

type ProfileHandler struct {
	profileUC      ProfileUseCase
	friendsUseCase FriendsUseCase
	authUseCase    AuthUseCase
	chatUseCase    ChatUseCase
	connService    IWebSocketManager
}

func NewProfileHandler(profileUC ProfileUseCase, friendUseCase FriendsUseCase, authUseCase AuthUseCase, chatUseCase ChatUseCase, connService IWebSocketManager) *ProfileHandler {
	return &ProfileHandler{
		profileUC:      profileUC,
		connService:    connService,
		friendsUseCase: friendUseCase,
		authUseCase:    authUseCase,
		chatUseCase:    chatUseCase,
	}
}

// GetProfile returns user profile
// @Summary Get user profile
// @Description Get user profile by id
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} forms.ProfileForm "User profile"
// @Failure 400 {object} forms.ErrorForm "Failed to parse user id"
// @Failure 404 {object} forms.ErrorForm "Profile not found"
// @Failure 500 {object} forms.ErrorForm "Failed to get profile"
// @Router /api/profile/{username} [get]
func (p *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// user whose profile is requested
	ctx := r.Context()
	userRequested := mux.Vars(r)["username"]
	logger.Info(ctx, fmt.Sprintf("Request profile of %s", userRequested))

	profileInfo, err := p.profileUC.GetUserInfoByUserName(ctx, userRequested)
	if errors.Is(err, usecase.ErrNotFound) {
		logger.Info(ctx, fmt.Sprintf("Profile of %s not found", userRequested))
		http2.WriteJSONError(w, "profile not found", http.StatusNotFound)
		return
	} else if err != nil {
		logger.Info(ctx, fmt.Sprintf("Unexpected error: %s", err.Error()))
		http2.WriteJSONError(w, "error while getting profile", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Profile of %s was successfully fetched", userRequested))

	_, isOnline := p.connService.IsConnected(profileInfo.UserId)

	var relation = models.RelationNone
	var chatId *uuid.UUID
	if session, err := r.Cookie("session"); err == nil {
		// parse session
		sessionUuid, err := uuid.Parse(session.Value)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to parse session: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to parse session", http.StatusBadRequest)
			return
		}

		// lookup user by session
		user, err := p.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to lookup user by session: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to authorize user", http.StatusUnauthorized)
			return
		}

		rel, err := p.friendsUseCase.GetUserRelation(ctx, user.Id, profileInfo.UserId)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to get user relation: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get user relation", http.StatusInternalServerError)
			return
		}
		relation = rel

		// get chat id
		chat, err := p.chatUseCase.GetPrivateChat(ctx, user.Id, profileInfo.UserId)
		if err != nil && !errors.Is(err, usecase.ErrNotFound) {
			logger.Error(ctx, fmt.Sprintf("Failed to get chat id: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get chat id", http.StatusInternalServerError)
			return
		} else {
			if err == nil {
				chatId = &chat.ID
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.ModelToForm(profileInfo, userRequested, isOnline, relation, chatId))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode profile: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode feed", http.StatusInternalServerError)
		return
	}
}

// UpdateProfile updates user profile
// @Summary Update user profile
// @Description Update user profile by id
// @Tags Profile
// @Accept json
// @Produce json
// @Param firstname formData string true "First name"
// @Param lastname formData string true "Last name"
// @Param birth_date formData string true "Birth date"
// @Param sex formData int true "Sex"
// @Param bio formData string true "Bio"
// @Param avatar formData file false "Avatar"
// @Success 200 {string} string "Profile updated"
// @Failure 400 {object} forms.ErrorForm "Failed to parse form"
// @Failure 500 {object} forms.ErrorForm "Failed to update profile"
// @Router /api/profile [post]
func (p *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating profile")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to update profile", user.Username))

	var profileForm forms.ProfileForm
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// retrieving files if passed
	profileForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get avatar: %s", err.Error()))
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get avatar: %v", err), http.StatusBadRequest)
		return
	}
	profileForm.Background, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get cover: %s", err.Error()))
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get cover: %v", err), http.StatusBadRequest)
		return
	}

	var recievedValidInfo = profileForm.Avatar != nil || profileForm.Background != nil
	// parsing main profile info
	var profileInfo forms.ProfileInfo
	err = json.NewDecoder(strings.NewReader(r.FormValue("profile"))).Decode(&profileInfo)
	if err == nil {
		profileForm.ProfileInfo = &profileInfo
		recievedValidInfo = true
	}

	// getting additional info
	var contactInfo forms.ContactInfo
	err = json.NewDecoder(strings.NewReader(r.FormValue("contact_info"))).Decode(&contactInfo)
	if err == nil {
		profileForm.ContactInfo = &contactInfo
		recievedValidInfo = true
	}

	var schoolEducation forms.SchoolEducationForm
	err = json.NewDecoder(strings.NewReader(r.FormValue("school"))).Decode(&schoolEducation)
	if err == nil {
		profileForm.SchoolEducation = &schoolEducation
		recievedValidInfo = true
	}

	var universityEducation forms.UniversityEducationForm
	err = json.NewDecoder(strings.NewReader(r.FormValue("university"))).Decode(&universityEducation)
	if err == nil {
		profileForm.UniversityEducation = &universityEducation
		recievedValidInfo = true
	}

	if !recievedValidInfo {
		logger.Error(ctx, "No valid data provided")
		http2.WriteJSONError(w, "No valid data provided", http.StatusBadRequest)
		return
	}

	// converting form to model
	profile, err := profileForm.FormToModel()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to convert form to model: %s", err.Error()))
		http2.WriteJSONError(w, fmt.Sprintf("Failed to parse form: %+v", err), http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Recieved profile update: %v", profile))

	profile.UserId = user.Id
	err = p.profileUC.UpdateProfile(ctx, profile)
	if errors.Is(err, usecase.ErrNotFound) {
		logger.Error(ctx, fmt.Sprintf("Profile of %s not found", user.Username))
		http2.WriteJSONError(w, "profile not found", http.StatusNotFound)
		return
	} else if errors.Is(err, usecase.ErrInvalidProfileInfo) {
		logger.Error(ctx, fmt.Sprintf("Invalid profile info: %s", err.Error()))
		http2.WriteJSONError(w, "invalid profile info", http.StatusBadRequest)
		return
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to update profile: %s", err.Error()))
		http2.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Profile of %s was successfully updated", user.Username))
}
