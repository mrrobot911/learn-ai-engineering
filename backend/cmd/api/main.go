package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/config"
	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/seed"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	seeddata "github.com/mrrobot911/learn-ai-engineering/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	h := handler.NewHealthHandler()

	var (
		contentRepository  repo.ContentRepository
		attemptRepository  repo.AttemptRepository
		authRepository     repo.AuthRepository
		progressRepository repo.ProgressRepository
	)

	if cfg.DatabaseURL != "" {
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer pool.Close()

		conn, err := pool.Acquire(ctx)
		if err != nil {
			log.Fatalf("Failed to acquire database connection: %v", err)
		}
		if err := migrate.Apply(ctx, conn.Conn(), os.DirFS("migrations")); err != nil {
			conn.Release()
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		conn.Release()

		seedReport, err := seed.Load(ctx, pool, seeddata.Files, os.Getenv("SEED_MAINTAINER_EMAIL"), os.Getenv("SEED_MAINTAINER_PASSWORD"))
		if err != nil {
			log.Fatalf("Failed to load seed content: %v", err)
		}
		log.Printf("Seed: %s", seedReport)

		contentRepository = repo.NewPostgresContentRepository(pool)
		attemptRepository = repo.NewPostgresAttemptRepository(pool)
		authRepository = repo.NewPostgresAuthRepository(pool)
		progressRepository = repo.NewPostgresProgressRepository(pool)
	} else {
		contentRepository = repo.NewMemoryContentRepository(nil, nil, nil)
		attemptRepository = repo.NewMemoryAttemptRepository()
		authRepository = repo.NewMemoryAuthRepository()
		progressRepository = repo.NewMemoryProgressRepository()
	}

	contentService := service.NewContentService(contentRepository, attemptRepository)
	contentHandler := handler.NewContentHandler(contentService)

	authService := service.NewAuthService(authRepository)
	authHandler := handler.NewAuthHandler(authService, cfg.Environment == "production")
	withSession := middleware.Session(authService)

	progressService := service.NewProgressService(progressRepository, contentRepository)
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
	mux.Handle("GET /me/attempts", withSession(http.HandlerFunc(contentHandler.ListMyAttempts)))
	mux.Handle("GET /me/attempts/{attemptId}", withSession(http.HandlerFunc(contentHandler.GetMyAttempt)))
	mux.Handle("PUT /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.MarkLessonComplete))))
	mux.Handle("DELETE /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.UnmarkLessonComplete))))
	mux.Handle("POST /topics/{topicSlug}/lessons", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.CreateDraftLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UpdateDraftLesson)))))
	mux.Handle("POST /topics/{topicSlug}/questions", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.CreateDraftQuestion)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UpdateDraftQuestion)))))
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}/publish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.PublishLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}/unpublish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UnpublishLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}/publish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.PublishQuestion)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}/unpublish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UnpublishQuestion)))))

	// Apply logging middleware
	loggedMux := middleware.Logging(handler.NewRouter(mux, cfg.StaticDir))

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
