package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/microcosm-cc/bluemonday"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"quickflow/config"
	qfhttp "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/middleware"
	"quickflow/gateway/internal/delivery/ws"
	friendsService "quickflow/shared/client/friends_service"
	"quickflow/shared/client/messenger_service"
	postService "quickflow/shared/client/post_service"
	userService "quickflow/shared/client/user_service"
)

const (
	filePort      = 8081
	postPort      = 8082
	userPort      = 8083
	messengerPort = 8084
	friendsPort   = 8085
)

func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	//grpcConnFileService, err := grpc.NewClient(
	//	fmt.Sprintf("localhost:%d", filePort),
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//)
	//if err != nil {
	//	return fmt.Errorf("failed to connect to file service: %w", err)
	//}

	grpcConnPostService, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", postPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to post service: %w", err)
	}

	grpcConnUserService, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", userPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %w", err)
	}

	grpcConnMessengerService, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", messengerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	grpcConnFriendsService, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", friendsPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// services
	//fileService := fileService.NewFileClient(grpcConnFileService)
	UserService := userService.NewUserServiceClient(grpcConnUserService)
	profileService := userService.NewProfileClient(grpcConnUserService)
	PostService := postService.NewPostServiceClient(grpcConnPostService)
	chatService := messenger_service.NewChatServiceClient(grpcConnMessengerService)
	messageService := messenger_service.NewMessageServiceClient(grpcConnMessengerService)
	friendsService := friendsService.NewFriendsClient(grpcConnFriendsService)

	connManager := ws.NewWSConnectionManager()
	sanitizerPolicy := bluemonday.UGCPolicy()

	newAuthHandler := qfhttp.NewAuthHandler(UserService, sanitizerPolicy)
	//newFeedHandler := qfhttp.NewFeedHandler(UserService, PostService, profileService, newFriendsService)
	newPostHandler := qfhttp.NewPostHandler(PostService, profileService, sanitizerPolicy)
	//newProfileHandler := qfhttp.NewProfileHandler(profileService, newFriendsService, UserService, chatService, connManager, sanitizerPolicy)
	newMessageHandler := qfhttp.NewMessageHandler(messageService, UserService, profileService, sanitizerPolicy)
	newChatHandler := qfhttp.NewChatHandler(chatService, profileService, connManager)
	newFriendsHandler := qfhttp.NewFriendsHandler(friendsService, connManager)
	newSearchHandler := qfhttp.NewSearchHandler(UserService)

	CSRFHandler := qfhttp.NewCSRFHandler()

	wsRouter := ws.NewWebSocketRouter()
	wsMessageHander := ws.NewInternalWSMessageHandler(connManager, messageService, profileService, chatService)
	pingHandler := ws.NewPingHandlerWS()

	// register handlers
	wsRouter.RegisterHandler("message", wsMessageHander.Handle)
	wsRouter.RegisterHandler("message_read", wsMessageHander.MarkMessageRead)

	newMessageHandlerWS := qfhttp.NewMessageListenerWS(profileService, connManager, wsRouter, sanitizerPolicy)

	// routing
	r := mux.NewRouter()
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.CORSMiddleware(cfg.CORSConfig))
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
	//r.HandleFunc("/profiles/{username}", newProfileHandler.GetProfile).Methods(http.MethodGet)

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json", "multipart/form-data"))

	apiGetRouter := r.PathPrefix("/").Subrouter()
	//apiGetRouter.HandleFunc("/profiles/{username}/posts", newFeedHandler.FetchUserPosts).Methods(http.MethodGet)

	// validating that the content type is application/json for every route but /hello

	apiPostRouter.HandleFunc("/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/login", newAuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/logout", newAuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(UserService))
	protectedPost.Use(middleware.CSRFMiddleware)
	protectedPost.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.UpdatePost).Methods(http.MethodPut)
	//protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/follow", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/followers/accept", newFriendsHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", newMessageHandler.SendMessageToUsername).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(UserService))
	//protectedGet.HandleFunc("/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	//protectedGet.HandleFunc("/recommendations", newFeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", CSRFHandler.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/users/search", newSearchHandler.SearchSimilar).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager, pingHandler))
	wsProtected.HandleFunc("/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(UserService))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/friends", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/follow", newFriendsHandler.Unfollow).Methods(http.MethodDelete)

	server := http.Server{
		Addr:         cfg.ServerConfig.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
	}

	fmt.Printf("starting server at %s\n", cfg.ServerConfig.Addr)
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}
