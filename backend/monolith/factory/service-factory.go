package factory

import (
	usecase2 "quickflow/monolith/internal/usecase"
)

type DefaultServiceFactory struct {
	repoFactory RepositoryFactory
}

func NewDefaultServiceFactory(repoFactory RepositoryFactory) *DefaultServiceFactory {
	return &DefaultServiceFactory{
		repoFactory: repoFactory,
	}
}

func (f *DefaultServiceFactory) AuthService() *usecase2.AuthService {
	return usecase2.NewAuthService(
		f.repoFactory.UserRepository(),
		f.repoFactory.SessionRepository(),
		f.repoFactory.ProfileRepository(),
	)
}

func (f *DefaultServiceFactory) ProfileService() *usecase2.ProfileService {
	return usecase2.NewProfileService(
		f.repoFactory.ProfileRepository(),
		f.repoFactory.UserRepository(),
		f.repoFactory.FileRepository(),
	)
}

func (f *DefaultServiceFactory) PostService() *usecase2.PostService {
	return usecase2.NewPostService(
		f.repoFactory.PostRepository(),
		f.repoFactory.FileRepository(),
	)
}

func (f *DefaultServiceFactory) ChatService() *usecase2.ChatService {
	return usecase2.NewChatUseCase(
		f.repoFactory.ChatRepository(),
		f.repoFactory.FileRepository(),
		f.repoFactory.ProfileRepository(),
		f.repoFactory.MessageRepository(),
	)
}

func (f *DefaultServiceFactory) MessageService() *usecase2.MessageService {
	return usecase2.NewMessageService(
		f.repoFactory.MessageRepository(),
		f.repoFactory.FileRepository(),
		f.repoFactory.ChatRepository(),
	)
}

func (f *DefaultServiceFactory) FriendService() *usecase2.FriendsService {
	return usecase2.NewFriendsService(
		f.repoFactory.FriendRepository(),
	)
}

func (f *DefaultServiceFactory) SearchService() *usecase2.SearchService {
	return usecase2.NewSearchService(
		f.repoFactory.UserRepository(),
	)
}
