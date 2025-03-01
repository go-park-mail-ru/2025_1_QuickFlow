package internal

import (
    "fmt"
    "net/http"
    quickflowhttp "quickflow/internal/delivery/http"
    "quickflow/internal/repository"
    "quickflow/internal/usecase"
    "time"
)

func Run() error {
    newRepo := repository.NewInMemory()
    newProcessor := usecase.NewProcessor(newRepo)
    newHandler := quickflowhttp.NewHandler(newProcessor)

    mux := http.NewServeMux()

    mux.HandleFunc("/hello", newHandler.Greet)

    server := http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    fmt.Println("starting server at :8080")
    err := server.ListenAndServe()
    if err != nil {
        return fmt.Errorf("internal.Run: %w", err)
    }

    return nil
}
