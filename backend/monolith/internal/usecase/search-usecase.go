package usecase

import (
	"context"
	"quickflow/monolith/internal/models"
)

type SearchService struct {
	userRepo UserRepository
}

func NewSearchService(userRepo UserRepository) *SearchService {
	return &SearchService{
		userRepo: userRepo,
	}
}

func (s *SearchService) SearchSimilarUser(ctx context.Context, toSearch string, postsCount uint) ([]models.PublicUserInfo, error) {
	users, err := s.userRepo.SearchSimilar(ctx, toSearch, postsCount)
	if err != nil {
		return nil, err
	}

	return users, nil
}
