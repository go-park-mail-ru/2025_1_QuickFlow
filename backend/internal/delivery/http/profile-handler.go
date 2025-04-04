package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	http2 "quickflow/utils/http"
)

type ProfileUseCase interface {
	GetUserInfoByUserName(ctx context.Context, username string) (models.Profile, error)
	UpdateProfile(ctx context.Context, newProfile models.Profile) error
}

type ProfileHandler struct {
	profileUC ProfileUseCase
}

func NewProfileHandler(profileUC ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		profileUC: profileUC,
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
// @Failure 404 {object} forms.ErrorForm "Failed to get profile"
// @Router /api/profile/{id} [get]
func (p *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// user whose profile is requested
	userRequested := mux.Vars(r)["username"]
	profileInfo, err := p.profileUC.GetUserInfoByUserName(r.Context(), userRequested)
	if err != nil {
		http2.WriteJSONError(w, "error while getting profile", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.ModelToForm(profileInfo, userRequested))
	if err != nil {
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
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var profileForm forms.ProfileForm
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	//profileForm.Name = r.FormValue("firstname")
	//profileForm.Surname = r.FormValue("lastname")
	//profileForm.Sex = models.Sex(sex)
	//profileForm.DateOfBirth = r.FormValue("birth_date")
	//profileForm.Bio = r.FormValue("bio")

	// retrieving files if passed
	profileForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get avatar: %v", err), http.StatusBadRequest)
		return
	}
	profileForm.Background, err = http2.GetFile(r, "cover")
	if err != nil {
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
		http2.WriteJSONError(w, "No valid data provided", http.StatusBadRequest)
		return
	}

	// converting form to model
	profile, err := profileForm.FormToModel()
	if err != nil {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	profile.UserId = user.Id
	err = p.profileUC.UpdateProfile(r.Context(), profile)
	if err != nil {
		http2.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
	}
}
