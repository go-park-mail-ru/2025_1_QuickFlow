package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/config"
	"quickflow/config/cors"
	minio_config "quickflow/config/minio"
	qfhttp "quickflow/internal/delivery/http"
	"quickflow/internal/delivery/http/middleware"
	"quickflow/internal/delivery/ws"
	"quickflow/internal/repository/minio"
	"quickflow/internal/repository/postgres"
	"quickflow/internal/repository/redis"
	"quickflow/internal/usecase"
)

func Run(cfg *config.Config, corsCfg *cors.CORSConfig, minioCfg *minio_config.MinioConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	newFileRepo, err := minio.NewMinioRepository(minioCfg)
	if err != nil {
		return fmt.Errorf("could not create minio repository: %v", err)
	}

	connManager := ws.NewWebSocketManager()

	sanitizerPolicy := bluemonday.UGCPolicy()

	newUserRepo := postgres.NewPostgresUserRepository()
	newPostRepo := postgres.NewPostgresPostRepository()
	newSessionRepo := redis.NewRedisSessionRepository()
	newProfileRepo := postgres.NewPostgresProfileRepository()
	newMessageRepo := postgres.NewPostgresMessageRepository()
	newChatRepo := postgres.NewPostgresChatRepository()
	newFriendsRepo := postgres.NewPostgresFriendsRepository()
	newAuthService := usecase.NewAuthService(newUserRepo, newSessionRepo, newProfileRepo)
	newPostService := usecase.NewPostService(newPostRepo, newFileRepo)
	newProfileService := usecase.NewProfileService(newProfileRepo, newUserRepo, newFileRepo)
	newMessageService := usecase.NewMessageUseCase(newMessageRepo, newFileRepo, newChatRepo)
	newChatService := usecase.NewChatUseCase(newChatRepo, newFileRepo, newProfileRepo, newMessageRepo)
	newFriendsService := usecase.NewFriendsService(newFriendsRepo)
	newSearchService := usecase.NewSearchService(newUserRepo)
	newAuthHandler := qfhttp.NewAuthHandler(newAuthService, sanitizerPolicy)
	newFeedHandler := qfhttp.NewFeedHandler(newAuthService, newPostService, newProfileService, newFriendsService)
	newPostHandler := qfhttp.NewPostHandler(newPostService, newProfileService, sanitizerPolicy)
	newProfileHandler := qfhttp.NewProfileHandler(newProfileService, newFriendsService, newAuthService, newChatService, connManager, sanitizerPolicy)
	newMessageHandler := qfhttp.NewMessageHandler(newMessageService, newAuthService, newProfileService, sanitizerPolicy)
	newChatHandler := qfhttp.NewChatHandler(newChatService, newProfileService, connManager)
	newFriendsHandler := qfhttp.NewFriendsHandler(newFriendsService, connManager)
	newSearchHandler := qfhttp.NewSearchHandler(newSearchService)

	defer newUserRepo.Close()
	defer newPostRepo.Close()
	defer newSessionRepo.Close()
	defer newProfileRepo.Close()
	defer newMessageRepo.Close()
	defer newChatRepo.Close()
	defer newFriendsRepo.Close()

	newMessageHandlerWS := qfhttp.NewMessageHandlerWS(newMessageService, newChatService, newProfileService, connManager, sanitizerPolicy)

	// routing
	r := mux.NewRouter()
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.CORSMiddleware(*corsCfg))
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.PathPrefix("/api/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}).Methods(http.MethodOptions)

	r.HandleFunc("/hello", newAuthHandler.Greet).Methods(http.MethodGet)
	r.HandleFunc("/profiles/{username}", newProfileHandler.GetProfile).Methods(http.MethodGet)

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json", "multipart/form-data"))

	apiGetRouter := r.PathPrefix("/").Subrouter()
	apiGetRouter.HandleFunc("/profiles/{username}/posts", newFeedHandler.FetchUserPosts).Methods(http.MethodGet)

	// validating that the content type is application/json for every route but /hello

	apiPostRouter.HandleFunc("/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/login", newAuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/logout", newAuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(newAuthService))
	protectedPost.Use(middleware.CSRFMiddleware)
	protectedPost.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.UpdatePost).Methods(http.MethodPut)
	protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/follow", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/followers/accept", newFriendsHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", newMessageHandler.SendMessageToUsername).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(newAuthService))
	protectedGet.HandleFunc("/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	protectedGet.HandleFunc("/recommendations", newFeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", qfhttp.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/users/search", newSearchHandler.SearchSimilar).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager))
	wsProtected.HandleFunc("/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(newAuthService))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/friends", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/follow", newFriendsHandler.Unfollow).Methods(http.MethodDelete)

	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	fmt.Printf("starting server at %s\n", cfg.Addr)
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}
