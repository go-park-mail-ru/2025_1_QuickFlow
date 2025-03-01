package internal

import (
	"fmt"
	"net/http"
	"quickflow/config"

	"github.com/gorilla/mux"

	qfhttp "quickflow/internal/delivery/http"
	"quickflow/internal/repository"
	"quickflow/internal/usecase"
)

func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	newRepo := repository.NewInMemory()
	newProcessor := usecase.NewProcessor(newRepo)
	newHandler := qfhttp.NewHandler(newProcessor)

	// routing
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.HandleFunc("/hello", newHandler.Greet).Methods(http.MethodGet)

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
