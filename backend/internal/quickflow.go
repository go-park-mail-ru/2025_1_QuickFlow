package internal

import (
    "fmt"
    "net/http"

    "github.com/gorilla/mux"

    "quickflow/config"
    qfhttp "quickflow/internal/delivery/http"
    "quickflow/internal/repository/in-memory"
    "quickflow/internal/repository/postgres_redis"
    "quickflow/internal/usecase"
)

func Run(cfg *config.Config) error {
    if cfg == nil {
        return fmt.Errorf("config is nil")
    }

    //newRepo := repository.NewInMemory()
    newUserRepo := postgres_redis.NewPostgresUserRepository()
    newPostRepo := postgres_redis.NewPostgresPostRepository()
    newSessionRepo := in_memory.NewInMemorySessionRepository()
    newAuthService := usecase.NewAuthService(newUserRepo, newSessionRepo)
    newPostService := usecase.NewPostService(newPostRepo)
    newAuthHandler := qfhttp.NewAuthHandler(newAuthService)
    newPostHandler := qfhttp.NewFeedHandler(newPostService, newAuthService)

    // routing
    r := mux.NewRouter()
    r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    })
    r.HandleFunc("/hello", newAuthHandler.Greet).Methods(http.MethodGet)
    r.HandleFunc("/feed", newPostHandler.GetFeed).Methods(http.MethodGet)
    r.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)
    r.HandleFunc("/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
    r.HandleFunc("/login", newAuthHandler.Login).Methods(http.MethodPost)

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
