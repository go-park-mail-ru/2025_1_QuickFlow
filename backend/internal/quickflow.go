package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"quickflow/config"
	qfhttp "quickflow/internal/delivery/http"
	"quickflow/internal/repository/in-memory"
	"quickflow/internal/usecase"
)

func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	//newRepo := repository.NewInMemory()
	newUserRepo := in_memory.NewInMemoryUserRepository()
	newPostRepo := in_memory.NewInMemoryPostRepository()
	newSessionRepo := in_memory.NewInMemorySessionRepository()
	newProcessor := usecase.NewProcessor(newUserRepo, newPostRepo, newSessionRepo)
	newHandler := qfhttp.NewHandler(newProcessor)

	// routing
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.HandleFunc("/hello", newHandler.Greet).Methods(http.MethodGet)
	r.HandleFunc("/feed", newHandler.GetFeed).Methods(http.MethodGet)
	r.HandleFunc("/post", newHandler.AddPost).Methods(http.MethodPost)
	r.HandleFunc("/signup", newHandler.SignUp).Methods(http.MethodPost)
	r.HandleFunc("/login", newHandler.Login).Methods(http.MethodPost)

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
