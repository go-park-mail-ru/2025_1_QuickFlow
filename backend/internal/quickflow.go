package internal

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"quickflow/config"
	qfhttp "quickflow/internal/delivery/http"
	"quickflow/internal/repository"
	"quickflow/internal/usecase"
)

func Run() error {
	newRepo := repository.NewInMemory()
	newProcessor := usecase.NewProcessor(newRepo)
	newHandler := qfhttp.NewHandler(newProcessor)

	// Supporting config path via flags
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Loading config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

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
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}
