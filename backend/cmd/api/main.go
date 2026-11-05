package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/config"
	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	h := handler.NewHealthHandler()
	contentRepository := repo.NewMemoryContentRepository(nil, nil, nil)
	contentService := service.NewContentService(contentRepository)
	contentHandler := handler.NewContentHandler(contentService)

	authService := service.NewAuthService(repo.NewMemoryAuthRepository())
	authHandler := handler.NewAuthHandler(authService, cfg.Environment == "production")
	withSession := middleware.Session(authService)

	progressService := service.NewProgressService(repo.NewMemoryProgressRepository(), contentRepository)
	progressHandler := handler.NewProgressHandler(progressService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /topics", contentHandler.ListTopics)
	mux.HandleFunc("GET /topics/{topicSlug}", contentHandler.GetTopic)
	mux.HandleFunc("GET /lessons/{lessonId}", contentHandler.GetLesson)
	mux.HandleFunc("POST /topics/{topicSlug}/quiz", contentHandler.StartQuiz)
	mux.Handle("POST /topics/{topicSlug}/quiz/submissions", withSession(middleware.OptionalCSRF(http.HandlerFunc(contentHandler.SubmitQuiz))))
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /auth/session", withSession(http.HandlerFunc(authHandler.CurrentSession)))
	mux.Handle("POST /auth/logout", withSession(middleware.CSRF(http.HandlerFunc(authHandler.Logout))))
	mux.Handle("GET /me/progress", withSession(http.HandlerFunc(progressHandler.GetMyProgress)))
	mux.Handle("PUT /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.MarkLessonComplete))))
	mux.Handle("DELETE /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.UnmarkLessonComplete))))

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
