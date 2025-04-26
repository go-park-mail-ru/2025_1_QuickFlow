package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"

	"quickflow/config"
	"quickflow/factory"
	"quickflow/internal/delivery/http/middleware"
)

func Run(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	repoFactory, err := factory.NewPGMFactory(config)
	if err != nil {
		return fmt.Errorf("could not create repositories: %v", err)
	}
	defer repoFactory.Close()

	// pattern abstract factory
	serviceFactory := factory.NewDefaultServiceFactory(repoFactory)
	handlerFactory := factory.NewHttpWSHandlerFactory(serviceFactory)

	handlers := handlerFactory.InitHttpHandlers()
	wsHandlers := handlerFactory.InitWSHandlers()

	r, err := setupRouters(config, handlers, wsHandlers, serviceFactory)
	if err != nil {
		return fmt.Errorf("could not setup routers: %v", err)
	}

	server := http.Server{
		Addr:         config.ServerConfig.Addr,
		Handler:      r,
		ReadTimeout:  config.ServerConfig.ReadTimeout,
		WriteTimeout: config.ServerConfig.WriteTimeout,
	}

	fmt.Printf("starting server at %s\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}

func setupRouters(cfg *config.Config, httpHandlers *factory.HttpHandlerCollection, wsHandlers *factory.WSHandlerCollection, serviceFactory factory.ServiceFactory) (*mux.Router, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

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

	r.HandleFunc("/hello", httpHandlers.AuthHandler.Greet).Methods(http.MethodGet)
	r.HandleFunc("/profiles/{username}", httpHandlers.ProfileHandler.GetProfile).Methods(http.MethodGet)

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json", "multipart/form-data"))

	apiGetRouter := r.PathPrefix("/").Subrouter()
	apiGetRouter.HandleFunc("/profiles/{username}/posts", httpHandlers.FeedHandler.FetchUserPosts).Methods(http.MethodGet)

	apiPostRouter.HandleFunc("/signup", httpHandlers.AuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/login", httpHandlers.AuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/logout", httpHandlers.AuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(serviceFactory.AuthService()))
	protectedPost.Use(middleware.CSRFMiddleware)
	protectedPost.HandleFunc("/post", httpHandlers.PostHandler.AddPost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", httpHandlers.PostHandler.UpdatePost).Methods(http.MethodPut)
	protectedPost.HandleFunc("/profile", httpHandlers.ProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/follow", httpHandlers.FriendHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/followers/accept", httpHandlers.FriendHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/users/{username:[0-9a-zA-Z-]+}/message", httpHandlers.MessageHandler.SendMessageToUsername).Methods(http.MethodPost)
	protectedPost.HandleFunc("/feedback", httpHandlers.FeedbackHandler.SaveFeedback).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(serviceFactory.AuthService()))
	protectedGet.HandleFunc("/feed", httpHandlers.FeedHandler.GetFeed).Methods(http.MethodGet)
	protectedGet.HandleFunc("/recommendations", httpHandlers.FeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", httpHandlers.MessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/chats", httpHandlers.ChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/friends", httpHandlers.FriendHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/csrf", httpHandlers.CSRFHandler.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/users/search", httpHandlers.SearchHandler.SearchSimilar).Methods(http.MethodGet)
	protectedGet.HandleFunc("/feedback", httpHandlers.FeedbackHandler.GetAllFeedbackType).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(wsHandlers.ConnManager, wsHandlers.PingHandler))
	wsProtected.HandleFunc("/ws", wsHandlers.MessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(serviceFactory.AuthService()))
	apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/posts/{post_id:[0-9a-fA-F-]{36}}", httpHandlers.PostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/friends", httpHandlers.FriendHandler.DeleteFriend).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/follow", httpHandlers.FriendHandler.Unfollow).Methods(http.MethodDelete)

	wsHandlers.WSRouter.RegisterHandler("message", wsHandlers.InternalWSMessageHandler.Handle)
	wsHandlers.WSRouter.RegisterHandler("message_read", wsHandlers.InternalWSMessageHandler.MarkMessageRead)

	return r, nil
}
