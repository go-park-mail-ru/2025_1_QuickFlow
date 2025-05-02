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
	micro_addr "quickflow/config/micro-addr"
	qfhttp "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/middleware"
	"quickflow/gateway/internal/delivery/ws"
	"quickflow/shared/client/feedback_service"
	"quickflow/shared/client/messenger_service"
	postService "quickflow/shared/client/post_service"
	userService "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	get_env "quickflow/utils/get-env"
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
		get_env.GetServiceAddr(micro_addr.DefaultPostServiceAddrEnv, micro_addr.DefaultPostServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to post service: %w", err)
	}

	grpcConnUserService, err := grpc.NewClient(
		get_env.GetServiceAddr(micro_addr.DefaultUserServiceAddrEnv, micro_addr.DefaultUserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %w", err)
	}

	grpcConnMessengerService, err := grpc.NewClient(
		get_env.GetServiceAddr(micro_addr.DefaultMessengerServiceAddrEnv, micro_addr.DefaultMessengerServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)

	grpcConnFeedbackService, err := grpc.NewClient(
		get_env.GetServiceAddr(micro_addr.DefaultFeedbackServiceAddrEnv, micro_addr.DefaultFeedbackServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)

	// services
	//fileService := fileService.NewFileClient(grpcConnFileService)
	UserService := userService.NewUserServiceClient(grpcConnUserService)
	profileService := userService.NewProfileClient(grpcConnUserService)
	PostService := postService.NewPostServiceClient(grpcConnPostService)
	chatService := messenger_service.NewChatServiceClient(grpcConnMessengerService)
	messageService := messenger_service.NewMessageServiceClient(grpcConnMessengerService)
	feedbackService := feedback_service.NewFeedbackClient(grpcConnFeedbackService)
	// TODO : friends

	connManager := ws.NewWSConnectionManager()
	sanitizerPolicy := bluemonday.UGCPolicy()

	newAuthHandler := qfhttp.NewAuthHandler(UserService, sanitizerPolicy)
	//newFeedHandler := qfhttp.NewFeedHandler(UserService, PostService, profileService, newFriendsService)
	newPostHandler := qfhttp.NewPostHandler(PostService, profileService, sanitizerPolicy)
	//newProfileHandler := qfhttp.NewProfileHandler(profileService, newFriendsService, UserService, chatService, connManager, sanitizerPolicy)
	newMessageHandler := qfhttp.NewMessageHandler(messageService, UserService, profileService, sanitizerPolicy)
	newChatHandler := qfhttp.NewChatHandler(chatService, profileService, connManager)
	//newFriendsHandler := qfhttp.NewFriendsHandler(newFriendsService, connManager)
	newSearchHandler := qfhttp.NewSearchHandler(UserService)
	CSRFHandler := qfhttp.NewCSRFHandler()
	FeedbackHandler := qfhttp.NewFeedbackHandler(feedbackService, profileService, sanitizerPolicy)

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
	protectedPost.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}/like", newPostHandler.LikePost).Methods(http.MethodPost)
	//protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	//protectedPost.HandleFunc("/follow", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	//protectedPost.HandleFunc("/followers/accept", newFriendsHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", newMessageHandler.SendMessageToUsername).Methods(http.MethodPost)
	protectedPost.HandleFunc("/feedback", FeedbackHandler.SaveFeedback).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(UserService))
	//protectedGet.HandleFunc("/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	//protectedGet.HandleFunc("/recommendations", newFeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	//protectedGet.HandleFunc("/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", CSRFHandler.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/users/search", newSearchHandler.SearchSimilar).Methods(http.MethodGet)
	protectedGet.HandleFunc("/feedback", FeedbackHandler.GetAllFeedbackType).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager, pingHandler))
	wsProtected.HandleFunc("/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(UserService))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}/like", newPostHandler.UnlikePost).Methods(http.MethodDelete)
	//apiDeleteRouter.HandleFunc("/friends", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	//apiDeleteRouter.HandleFunc("/follow", newFriendsHandler.Unfollow).Methods(http.MethodDelete)

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
