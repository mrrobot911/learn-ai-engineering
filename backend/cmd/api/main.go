package main

import (
	"log"
	"net/http"

	"github.com/datatalksclub/learn-ai-engineering/internal/config"
	"github.com/datatalksclub/learn-ai-engineering/internal/handler"
	"github.com/datatalksclub/learn-ai-engineering/internal/middleware"
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
	if err := http.ListenAndServe(cfg.Addr(), loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
