package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"quickflow/gateway/internal/delivery/forms"
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
}

func NewSearchHandler(searchUseCase SearchUseCase, communityService CommunityService) *SearchHandler {
	return &SearchHandler{
		searchUseCase:    searchUseCase,
		communityService: communityService,
	}
}

func (s *SearchHandler) SearchSimilarUsers(w http.ResponseWriter, r *http.Request) {
	var searchForm forms.SearchForm
	err := searchForm.Unpack(r.URL.Query())
	if err != nil {
		logger.Error(r.Context(), "Failed to decode request body for user search: "+err.Error())
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	users, err := s.searchUseCase.SearchSimilarUser(r.Context(), searchForm.ToSearch, searchForm.Count)
	if err != nil {
		err := errors.FromGRPCError(err)
		logger.Error(r.Context(), fmt.Sprintf("Failed to search similar users: %s", err.Error()))
		http.Error(w, "Failed to search similar users", err.HTTPStatus)
		return
	}

	var publicUsersInfoOut []forms.PublicUserInfoOut
	for _, user := range users {
		publicUsersInfoOut = append(publicUsersInfoOut, forms.PublicUserInfoToOut(user, ""))
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.PublicUserInfoOut]{Payload: publicUsersInfoOut})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *SearchHandler) SearchSimilarCommunities(w http.ResponseWriter, r *http.Request) {
	var searchForm forms.SearchForm
	err := searchForm.Unpack(r.URL.Query())
	if err != nil {
		logger.Error(r.Context(), "Failed to decode request body for user search: "+err.Error())
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	communities, err := s.communityService.SearchSimilarCommunities(r.Context(), searchForm.ToSearch, int(searchForm.Count))
	if err != nil {
		err := errors.FromGRPCError(err)
		logger.Error(r.Context(), fmt.Sprintf("Failed to search similar communities: %s", err.Error()))
		http.Error(w, "Failed to search similar communities", err.HTTPStatus)
		return
	}

	out := make([]forms.CommunityForm, len(communities))
	for i, community := range communities {
		out[i] = forms.ToCommunityForm(*community)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[[]forms.CommunityForm]{Payload: out})
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Failed to encode user communities: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode user communities", http.StatusInternalServerError)
		return
	}
}
