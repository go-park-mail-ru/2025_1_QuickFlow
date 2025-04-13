package internal

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
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
	newAuthHandler := qfhttp.NewAuthHandler(newAuthService)
	newFeedHandler := qfhttp.NewFeedHandler(newPostService, newProfileService)
	newPostHandler := qfhttp.NewPostHandler(newPostService)
	newProfileHandler := qfhttp.NewProfileHandler(newProfileService, connManager)
	newMessageHandler := qfhttp.NewMessageHandler(newMessageService, newAuthService, newProfileService)
	newChatHandler := qfhttp.NewChatHandler(newChatService, newProfileService, connManager)
	newFriendsHandler := qfhttp.NewFriendsHandler(newFriendsService, connManager)

	defer newUserRepo.Close()
	defer newPostRepo.Close()
	defer newSessionRepo.Close()
	defer newProfileRepo.Close()
	defer newMessageRepo.Close()
	defer newChatRepo.Close()
	defer newFriendsRepo.Close()

	newMessageHandlerWS := qfhttp.NewMessageHandlerWS(newMessageService, newChatService, newProfileService, connManager)

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
	// validating that the content type is application/json for every route but /hello

	apiPostRouter.HandleFunc("/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/login", newAuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/logout", newAuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(newAuthService))
	protectedPost.Use(middleware.CSRFMiddleware)
	protectedPost.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/friends", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", newMessageHandler.SendMessageToUsername).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(newAuthService))
	protectedGet.HandleFunc("/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", qfhttp.GetCSRF).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager))
	wsProtected.HandleFunc("/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(newAuthService))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)

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
