package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/pkg/logger"
)

type SearchUseCase interface {
	SearchSimilarUser(ctx context.Context, toSearch string, postsCount uint) ([]models.PublicUserInfo, error)
}

type SearchHandler struct {
	searchUseCase SearchUseCase
}

func NewSearchHandler(searchUseCase SearchUseCase) *SearchHandler {
	return &SearchHandler{
		searchUseCase: searchUseCase,
	}
}

func (s *SearchHandler) SearchSimilar(w http.ResponseWriter, r *http.Request) {
	var searchForm forms.SearchForm
	err := searchForm.Unpack(r.URL.Query())
	if err != nil {
		logger.Error(r.Context(), "Failed to decode request body for user search: "+err.Error())
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	users, err := s.searchUseCase.SearchSimilarUser(r.Context(), searchForm.ToSearch, searchForm.UsersCount)
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Failed to search similar users: %s", err.Error()))
		http.Error(w, "Failed to search similar users", http.StatusInternalServerError)
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
