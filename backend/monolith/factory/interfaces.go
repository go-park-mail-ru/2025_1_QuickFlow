package factory

import (
	usecase2 "quickflow/monolith/internal/usecase"
)

type RepositoryFactory interface {
	UserRepository() usecase2.UserRepository
	SessionRepository() usecase2.SessionRepository
	ProfileRepository() usecase2.ProfileRepository
	PostRepository() usecase2.PostRepository
	ChatRepository() usecase2.ChatRepository
	MessageRepository() usecase2.MessageRepository
	FileRepository() usecase2.FileRepository
	FriendRepository() usecase2.FriendsRepository
	Close() error
}

type ServiceFactory interface {
	AuthService() *usecase2.AuthService
	ProfileService() *usecase2.ProfileService
	PostService() *usecase2.PostService
	ChatService() *usecase2.ChatService
	MessageService() *usecase2.MessageService
	FriendService() *usecase2.FriendsService
	SearchService() *usecase2.SearchService
}

type HandlerFactory interface {
	InitHttpHandlers() *HttpHandlerCollection
	InitWSHandlers() *WSHandlerCollection
}
