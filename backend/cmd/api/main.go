package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/config"
	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	h := handler.NewHealthHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)

	// Apply logging middleware
	loggedMux := middleware.Logging(mux)

	log.Printf("Starting server on %s", cfg.Addr())
	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           loggedMux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
