package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/forms"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type CommunityService interface {
	CreateCommunity(ctx context.Context, community *models.Community) (*models.Community, error)
	GetCommunityById(ctx context.Context, id uuid.UUID) (*models.Community, error)
	GetCommunityByName(ctx context.Context, name string) (*models.Community, error)
	GetCommunityMembers(ctx context.Context, communityId uuid.UUID, count int, ts time.Time) ([]*models.CommunityMember, error)
	IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error)
	DeleteCommunity(ctx context.Context, communityId uuid.UUID, userId uuid.UUID) error
	UpdateCommunity(ctx context.Context, community *models.Community, userId uuid.UUID) (*models.Community, error)
	JoinCommunity(ctx context.Context, member *models.CommunityMember) error
	LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error
	GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]*models.Community, error)
	SearchSimilarCommunities(ctx context.Context, name string, count int) ([]*models.Community, error)
	ChangeUserRole(ctx context.Context, userId, communityId uuid.UUID, role models.CommunityRole, requester uuid.UUID) error
	GetControlledCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]*models.Community, error)
}

type CommunityHandler struct {
	communityService CommunityService
	profileService   ProfileUseCase
	connService      IWebSocketConnectionManager
	authService      AuthUseCase
	policy           *bluemonday.Policy
}

func NewCommunityHandler(communityService CommunityService, profileService ProfileUseCase, connService IWebSocketConnectionManager, authService AuthUseCase, policy *bluemonday.Policy) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
		profileService:   profileService,
		connService:      connService,
		authService:      authService,
		policy:           policy,
	}
}

func (c *CommunityHandler) CreateCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while creating community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community creation", user.Username))

	err := r.ParseMultipartForm(15 << 20) // 10 MB
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// parsing JSON
	var communityForm forms.CreateCommunityForm
	communityForm.Nickname = r.FormValue("nickname")
	communityForm.Name = r.FormValue("name")
	communityForm.Description = r.FormValue("description")

	if utf8.RuneCountInString(communityForm.Description) > 500 {
		logger.Error(ctx, fmt.Sprintf("Text length validation failed: length=%d", utf8.RuneCountInString(communityForm.Description)))
		http2.WriteJSONError(w, "Text must be between 1 and 4096 characters", http.StatusBadRequest)
		return
	}

	sanitizer.SanitizeCommunityCreation(&communityForm, c.policy)

	communityForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get avatar: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get avatar", http.StatusBadRequest)
		return
	}

	communityForm.Cover, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get cover: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get cover", http.StatusBadRequest)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Recieved community: %+v", communityForm))

	community := communityForm.CreateFormToModel()
	community.OwnerID = user.Id

	newCommunity, err := c.communityService.CreateCommunity(ctx, &community)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to create community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to create community", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully added community")

	// owner public info
	info, err := c.profileService.GetPublicUserInfo(ctx, newCommunity.OwnerID)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
	}

	// Create response
	communityOut := forms.ToCommunityForm(*newCommunity, info)
	communityOut.Role = string(models.CommunityRoleOwner)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.CommunityForm]{Payload: communityOut})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode community", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) GetCommunityById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community info", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	community, err := c.communityService.GetCommunityById(ctx, communityId)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get community", err.HTTPStatus)
		return
	}

	isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, communityId)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to check community membership: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to check community membership", err.HTTPStatus)
		return
	}

	info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
	}

	out := forms.ToCommunityForm(*community, info)
	if isMember && role != nil {
		out.Role = string(*role)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode community", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) GetCommunityByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community info", user.Username))

	communityName := mux.Vars(r)["name"]
	if len(communityName) == 0 {
		http2.WriteJSONError(w, "Community name is required", http.StatusBadRequest)
		return
	}

	community, err := c.communityService.GetCommunityByName(ctx, communityName)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get community", err.HTTPStatus)
		return
	}

	isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to check community membership: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to check community membership", err.HTTPStatus)
		return
	}

	info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
	}
	out := forms.ToCommunityForm(*community, info)
	if isMember && role != nil {
		out.Role = string(*role)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode community", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) GetCommunityMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community members")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community members", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse query params: %v", err))
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	members, err := c.communityService.GetCommunityMembers(ctx, communityId, pagination.Count, pagination.Ts)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get community members: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get community members", err.HTTPStatus)
		return
	}

	memberIds := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		memberIds = append(memberIds, member.UserID)
	}
	var out []forms.CommunityMemberOut
	publicInfos, err := c.profileService.GetPublicUsersInfo(ctx, memberIds)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get public user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get public user info", http.StatusInternalServerError)
		return
	}

	publicInfoMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicInfos {
		publicInfoMap[info.Id] = info
	}
	for _, member := range members {
		formOut := forms.ToCommunityMemberOut(*member, publicInfoMap[member.UserID])
		_, isOnline := c.connService.IsConnected(member.UserID)
		if isOnline {
			formOut.IsOnline = &isOnline
		} else {
			formOut.IsOnline = nil
		}
		out = append(out, formOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.CommunityMemberOut]{Payload: out})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode community members: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode community members", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) DeleteCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while deleting community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community deletion", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	err = c.communityService.DeleteCommunity(ctx, communityId, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to delete community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to delete community", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully deleted community")
}

func (c *CommunityHandler) UpdateCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested community update", user.Username))

	err := r.ParseMultipartForm(15 << 20) // 10 MB
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	// parsing JSON
	var communityForm forms.CreateCommunityForm
	communityForm.Nickname = r.FormValue("nickname")
	communityForm.Name = r.FormValue("name")
	communityForm.Description = r.FormValue("description")

	if utf8.RuneCountInString(communityForm.Description) > 4000 {
		logger.Error(ctx, fmt.Sprintf("Text length validation failed: length=%d", utf8.RuneCountInString(communityForm.Description)))
		http2.WriteJSONError(w, "Text must be between 1 and 4096 characters", http.StatusBadRequest)
		return
	}

	sanitizer.SanitizeCommunityCreation(&communityForm, c.policy)

	communityForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get avatar: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get avatar", http.StatusBadRequest)
		return
	}

	communityForm.Cover, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get cover: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get cover", http.StatusBadRequest)
		return
	}

	community := communityForm.CreateFormToModel()
	community.ID = communityId

	// getting additional info
	var contactInfo forms.ContactInfo
	err = json.NewDecoder(strings.NewReader(r.FormValue("contact_info"))).Decode(&contactInfo)
	if err == nil {
		community.ContactInfo = forms.ContactInfoFormToModel(&contactInfo)
	}

	logger.Info(ctx, fmt.Sprintf("Recieved community: %+v", communityForm))

	newCommunity, err := c.communityService.UpdateCommunity(ctx, &community, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to update community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to update community", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully updated community")

	info, err := c.profileService.GetPublicUserInfo(ctx, newCommunity.OwnerID)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.CommunityForm]{Payload: forms.ToCommunityForm(*newCommunity, info)})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode community", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) JoinCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while joining community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to join community", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	member := models.CommunityMember{
		UserID:      user.Id,
		CommunityID: communityId,
		// TODO
		Role: models.CommunityRoleMember,
	}

	err = c.communityService.JoinCommunity(ctx, &member)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to join community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to join community", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully joined community")
}

func (c *CommunityHandler) LeaveCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while leaving community")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to leave community", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	err = c.communityService.LeaveCommunity(ctx, user.Id, communityId)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to leave community: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to leave community", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully left community")
}

func (c *CommunityHandler) GetUserCommunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := mux.Vars(r)["username"]
	if username == "" {
		http2.WriteJSONError(w, "Failed to get username from URL", http.StatusBadRequest)
		return
	}

	user, err := c.authService.GetUserByUsername(ctx, username)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get user: %s", err.Error()), err.HTTPStatus)
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse query params: %v", err))
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	communities, err := c.communityService.GetUserCommunities(ctx, user.Id, pagination.Count, pagination.Ts)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user communities: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user communities", err.HTTPStatus)
		return
	}

	out := make([]forms.CommunityForm, len(communities))
	for i, community := range communities {
		info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
		}

		out[i] = forms.ToCommunityForm(*community, info)
		isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to check community membership: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to check community membership", err.HTTPStatus)
			return
		}
		if isMember && role != nil {
			out[i].Role = string(*role)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode user communities: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode user communities", http.StatusInternalServerError)
		return
	}
}

func (c *CommunityHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while changing community role")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to change community role", user.Username))

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, "Community ID is required", http.StatusBadRequest)
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Invalid community ID: %s", err.Error()))
		http2.WriteJSONError(w, "Invalid community ID", http.StatusBadRequest)
		return
	}

	userId := mux.Vars(r)["user_id"]
	if len(userId) == 0 {
		logger.Error(ctx, "User ID is required")
		http2.WriteJSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Invalid user ID: %s", err.Error()))
		http2.WriteJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	role := r.URL.Query().Get("role")
	if len(role) == 0 {
		logger.Error(ctx, "Role is required")
		http2.WriteJSONError(w, "Role is required", http.StatusBadRequest)
		return
	}

	switch role {
	case "member":
		role = string(models.CommunityRoleMember)
	case "admin":
		role = string(models.CommunityRoleAdmin)
	case "owner":
		role = string(models.CommunityRoleOwner)
	default:
		logger.Error(ctx, "Invalid role provided", http.StatusBadRequest)
		http2.WriteJSONError(w, "Invalid role", http.StatusBadRequest)
		return
	}

	roleParsed := models.CommunityRole(role)

	err = c.communityService.ChangeUserRole(ctx, userIdParsed, communityId, roleParsed, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to change community role: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to change community role", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully changed community role")
}

func (c *CommunityHandler) GetControlledCommunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := mux.Vars(r)["username"]
	if username == "" {
		http2.WriteJSONError(w, "Failed to get username from URL", http.StatusBadRequest)
		return
	}

	user, err := c.authService.GetUserByUsername(ctx, username)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get user: %s", err.Error()), err.HTTPStatus)
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse query params: %v", err))
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	communities, err := c.communityService.GetControlledCommunities(ctx, user.Id, pagination.Count, pagination.Ts)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get user communities: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get user communities", err.HTTPStatus)
		return
	}

	out := make([]forms.CommunityForm, len(communities))
	for i, community := range communities {
		info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
		}

		out[i] = forms.ToCommunityForm(*community, info)
		isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to check community membership: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to check community membership", err.HTTPStatus)
			return
		}
		if isMember && role != nil {
			out[i].Role = string(*role)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode user communities: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode user communities", http.StatusInternalServerError)
		return
	}
}
