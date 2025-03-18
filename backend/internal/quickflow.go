package internal

import (
	"fmt"
	"net/http"
	"quickflow/internal/repository/redis"

	"github.com/gorilla/mux"

	"quickflow/config"
	qfhttp "quickflow/internal/delivery/http"
	"quickflow/internal/delivery/http/middleware"
	"quickflow/internal/repository/postgres"
	"quickflow/internal/usecase"
)

func Run(cfg *config.Config, corsCfg *config.CORSConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	//newRepo := repository.NewInMemory()
	newUserRepo := postgres.NewPostgresUserRepository()
	newPostRepo := postgres.NewPostgresPostRepository()
	newSessionRepo := redis.NewRedisSessionRepository()
	newAuthService := usecase.NewAuthService(newUserRepo, newSessionRepo)
	newPostService := usecase.NewPostService(newPostRepo)
	newAuthHandler := qfhttp.NewAuthHandler(newAuthService)
	newPostHandler := qfhttp.NewFeedHandler(newPostService, newAuthService)
	defer newPostRepo.Close()
	defer newSessionRepo.Close()

	// routing
	r := mux.NewRouter()
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

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json"))

	apiGetRouter := r.PathPrefix("/").Subrouter()
	// validating that the content type is application/json for every route but /hello

	apiPostRouter.HandleFunc("/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/login", newAuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/logout", newAuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(newAuthService))
	protectedPost.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(newAuthService))
	protectedGet.HandleFunc("/feed", newPostHandler.GetFeed).Methods(http.MethodGet)

	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	fmt.Printf("starting server at %s\n", cfg.Addr)
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}
