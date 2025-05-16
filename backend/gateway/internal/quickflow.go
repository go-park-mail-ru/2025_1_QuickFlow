package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/microcosm-cc/bluemonday"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"quickflow/config"
	addr "quickflow/config/micro-addr"
	qfhttp "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/middleware"
	"quickflow/gateway/internal/delivery/ws"
	"quickflow/metrics"
	"quickflow/shared/client/community_service"
	"quickflow/shared/client/feedback_service"
	"quickflow/shared/client/file_service"
	friendsService "quickflow/shared/client/friends_service"
	"quickflow/shared/client/messenger_service"
	postService "quickflow/shared/client/post_service"
	userService "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	getEnv "quickflow/utils/get-env"
)

func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	metrics := metrics.NewMetrics("QuickFlow")

	grpcConnFileService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFileServiceAddrEnv, addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	grpcConnPostService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultPostServiceAddrEnv, addr.DefaultPostServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to post service: %w", err)
	}

	grpcConnUserService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultUserServiceAddrEnv, addr.DefaultUserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %w", err)
	}

	grpcConnMessengerService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultMessengerServiceAddrEnv, addr.DefaultMessengerServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	grpcConnFeedbackService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFeedbackServiceAddrEnv, addr.DefaultFeedbackServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	grpcConnFriendsService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFriendsServiceAddrEnv, addr.DefaultFriendsServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	grcpConnCommunityService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultCommunityServiceAddrEnv, addr.DefaultCommunityServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	// services
	fileService := file_service.NewFileClient(grpcConnFileService)
	UserService := userService.NewUserServiceClient(grpcConnUserService)
	profileService := userService.NewProfileClient(grpcConnUserService)
	PostService := postService.NewPostServiceClient(grpcConnPostService)
	chatService := messenger_service.NewChatServiceClient(grpcConnMessengerService)
	messageService := messenger_service.NewMessageServiceClient(grpcConnMessengerService)
	feedbackService := feedback_service.NewFeedbackClient(grpcConnFeedbackService)
	FriendsService := friendsService.NewFriendsClient(grpcConnFriendsService)
	communityService := community_service.NewCommunityServiceClient(grcpConnCommunityService)

	connManager := ws.NewWSConnectionManager()
	sanitizerPolicy := bluemonday.UGCPolicy()

	newAuthHandler := qfhttp.NewAuthHandler(UserService, sanitizerPolicy)
	newFeedHandler := qfhttp.NewFeedHandler(UserService, PostService, profileService, FriendsService, communityService)
	newPostHandler := qfhttp.NewPostHandler(PostService, profileService, communityService, FriendsService, sanitizerPolicy)
	newProfileHandler := qfhttp.NewProfileHandler(profileService, FriendsService, UserService, chatService, connManager, sanitizerPolicy)
	newMessageHandler := qfhttp.NewMessageHandler(messageService, UserService, profileService, sanitizerPolicy)
	newChatHandler := qfhttp.NewChatHandler(chatService, profileService, connManager)
	newFriendsHandler := qfhttp.NewFriendsHandler(FriendsService, connManager)
	newSearchHandler := qfhttp.NewSearchHandler(UserService, communityService, profileService)
	newCommunityHandler := qfhttp.NewCommunityHandler(communityService, profileService, connManager, UserService, sanitizerPolicy)
	newFileHandler := qfhttp.NewFileHandler(fileService, sanitizerPolicy)

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
	r.Use(middleware.MetricsMiddleware(metrics))
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
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json", "multipart/form-data"))

	apiGetRouter := r.PathPrefix("/").Subrouter()

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
	protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/follow", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/followers/accept", newFriendsHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", newMessageHandler.SendMessageToUsername).Methods(http.MethodPost)
	protectedPost.HandleFunc("/feedback", FeedbackHandler.SaveFeedback).Methods(http.MethodPost)
	protectedPost.HandleFunc("/community", newCommunityHandler.CreateCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.UpdateCommunity).Methods(http.MethodPut)
	protectedPost.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}/join", newCommunityHandler.JoinCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}/leave", newCommunityHandler.LeaveCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}/members/{user_id:{id:[0-9a-fA-F-]{36}}", newCommunityHandler.ChangeUserRole).Methods(http.MethodPost)
	protectedPost.HandleFunc("/upload", newFileHandler.AddFiles).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(UserService))
	protectedGet.HandleFunc("/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	protectedGet.HandleFunc("/recommendations", newFeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", CSRFHandler.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/users/search", newSearchHandler.SearchSimilarUsers).Methods(http.MethodGet)
	protectedGet.HandleFunc("/communities/search", newSearchHandler.SearchSimilarCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/feedback", FeedbackHandler.GetAllFeedbackType).Methods(http.MethodGet)
	protectedGet.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.GetCommunityById).Methods(http.MethodGet)
	protectedGet.HandleFunc("/communities/{name}", newCommunityHandler.GetCommunityByName).Methods(http.MethodGet)
	protectedGet.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}/members", newCommunityHandler.GetCommunityMembers).Methods(http.MethodGet)
	protectedGet.HandleFunc("/profiles/{username}/communities", newCommunityHandler.GetUserCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/profiles/{username}/controlled", newCommunityHandler.GetControlledCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/communities/{name}/posts", newFeedHandler.FetchCommunityPosts).Methods(http.MethodGet)
	protectedGet.HandleFunc("/profiles/{username}/posts", newFeedHandler.FetchUserPosts).Methods(http.MethodGet)
	protectedGet.HandleFunc("/my_profile", newProfileHandler.GetMyProfile).Methods(http.MethodGet)
	protectedGet.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.GetPost).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager, pingHandler))
	wsProtected.HandleFunc("/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(UserService))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}/like", newPostHandler.UnlikePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/friends", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/follow", newFriendsHandler.Unfollow).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.DeleteCommunity).Methods(http.MethodDelete)

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
