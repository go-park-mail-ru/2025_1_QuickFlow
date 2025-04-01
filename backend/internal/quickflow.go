package internal

import (
    "fmt"
    "net/http"
    "quickflow/config/cors"
    minio_config "quickflow/config/minio"
    "quickflow/internal/repository/minio"
    "quickflow/internal/repository/redis"

    "github.com/gorilla/mux"

    "quickflow/config"
    qfhttp "quickflow/internal/delivery/http"
    "quickflow/internal/delivery/http/middleware"
    "quickflow/internal/repository/postgres"
    "quickflow/internal/usecase"
)

func Run(cfg *config.Config, corsCfg *cors.CORSConfig, minioCfg *minio_config.MinioConfig) error {
    if cfg == nil {
        return fmt.Errorf("config is nil")
    }

    //newRepo := repository.NewInMemory()
    newFileRepo, err := minio.NewMinioRepository(minioCfg)
    if err != nil {
        return fmt.Errorf("could not create minio repository: %v", err)
    }
    newUserRepo := postgres.NewPostgresUserRepository()
    newPostRepo := postgres.NewPostgresPostRepository()
    newSessionRepo := redis.NewRedisSessionRepository()
    newProfileRepo := postgres.NewPostgresProfileRepository()
    newAuthService := usecase.NewAuthService(newUserRepo, newSessionRepo, newProfileRepo)
    newPostService := usecase.NewPostService(newPostRepo, newFileRepo)
    newProfileService := usecase.NewProfileService(newProfileRepo, newFileRepo)
    newAuthHandler := qfhttp.NewAuthHandler(newAuthService)
    newPostHandler := qfhttp.NewFeedHandler(newPostService, newAuthService)
    newProfileHandler := qfhttp.NewProfileHandler(newProfileService)
    defer newUserRepo.Close()
    defer newPostRepo.Close()
    defer newSessionRepo.Close()
    defer newProfileRepo.Close()

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
    r.HandleFunc("/profile/{id:[0-9a-fA-F-]{36}}", newProfileHandler.GetProfile).Methods(http.MethodGet)

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
    protectedPost.HandleFunc("/post", newPostHandler.AddPost).Methods(http.MethodPost)
    protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
    protectedGet.Use(middleware.SessionMiddleware(newAuthService))
    protectedGet.HandleFunc("/feed", newPostHandler.GetFeed).Methods(http.MethodGet)
    protectedPost.HandleFunc("/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)

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
