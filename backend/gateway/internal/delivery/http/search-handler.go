package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type SearchUseCase interface {
	SearchSimilarUser(ctx context.Context, toSearch string, postsCount uint) ([]models.PublicUserInfo, error)
}

type SearchHandler struct {
	searchUseCase    SearchUseCase
	communityService CommunityService
	profileService   ProfileUseCase
}

func NewSearchHandler(searchUseCase SearchUseCase, communityService CommunityService, profileService ProfileUseCase) *SearchHandler {
	return &SearchHandler{
		searchUseCase:    searchUseCase,
		communityService: communityService,
		profileService:   profileService,
	}
}

func (s *SearchHandler) SearchSimilarUsers(w http.ResponseWriter, r *http.Request) {
	var searchForm forms.SearchForm
	err := searchForm.Unpack(r.URL.Query())
	if err != nil {
		logger.Error(r.Context(), "Failed to decode request body for user search: "+err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to decode request body", http.StatusBadRequest))
		return
	}

	users, err := s.searchUseCase.SearchSimilarUser(r.Context(), searchForm.ToSearch, searchForm.Count)
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Failed to search similar users: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	var publicUsersInfoOut []forms.PublicUserInfoOut
	for _, user := range users {
		publicUsersInfoOut = append(publicUsersInfoOut, forms.PublicUserInfoToOut(user, ""))
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.PublicUserInfoOut]{Payload: publicUsersInfoOut})
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode user communities", http.StatusInternalServerError))
		return
	}
}

func (s *SearchHandler) SearchSimilarCommunities(w http.ResponseWriter, r *http.Request) {
	var searchForm forms.SearchForm
	err := searchForm.Unpack(r.URL.Query())
	if err != nil {
		logger.Error(r.Context(), "Failed to decode request body for user search: "+err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to decode request body", http.StatusBadRequest))
		return
	}

	communities, err := s.communityService.SearchSimilarCommunities(r.Context(), searchForm.ToSearch, int(searchForm.Count))
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Failed to search similar communities: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	out := make([]forms.CommunityForm, len(communities))
	for i, community := range communities {
		info, err := s.profileService.GetPublicUserInfo(r.Context(), community.OwnerID)
		if err != nil {
			logger.Error(r.Context(), fmt.Sprintf("Failed to get user info: %s", err.Error()))
			http2.WriteJSONError(w, err)
		}
		out[i] = forms.ToCommunityForm(*community, info)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Failed to encode user communities: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode user communities", http.StatusInternalServerError))
		return
	}
}
