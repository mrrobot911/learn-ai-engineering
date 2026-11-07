//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/seed"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	seeddata "github.com/mrrobot911/learn-ai-engineering/seed"
)

func integrationDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL required for integration tests")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}
	dbName := cfg.ConnConfig.Database
	if dbName == "learn_ai_engineering" || !strings.Contains(strings.ToLower(dbName), "test") {
		t.Fatalf("refusing to run integration tests against non-test database %q (set TEST_DATABASE_URL to an isolated test database)", dbName)
	}
	return url
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, integrationDBURL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func integrationMigrationsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

func resetIntegrationTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DROP TABLE IF EXISTS quiz_attempt_answers`,
		`DROP TABLE IF EXISTS quiz_attempts`,
		`DROP TABLE IF EXISTS lesson_progress`,
		`DROP TABLE IF EXISTS sessions`,
		`DROP TABLE IF EXISTS users`,
		`DROP TABLE IF EXISTS answer_options`,
		`DROP TABLE IF EXISTS questions`,
		`DROP TABLE IF EXISTS lessons`,
		`DROP TABLE IF EXISTS topics`,
		`DROP TABLE IF EXISTS schema_migrations`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			t.Fatalf("reset schema (%s): %v", s, err)
		}
	}
}

func applyIntegrationMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()
	if err := migrate.Apply(ctx, conn.Conn(), os.DirFS(integrationMigrationsDir(t))); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func buildIntegrationHandler(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	contentRepository := repo.NewPostgresContentRepository(pool)
	attemptRepository := repo.NewPostgresAttemptRepository(pool)
	authRepository := repo.NewPostgresAuthRepository(pool)
	progressRepository := repo.NewPostgresProgressRepository(pool)

	contentService := service.NewContentServiceWithRandom(contentRepository, attemptRepository, rand.New(rand.NewSource(1)))
	contentHandler := handler.NewContentHandler(contentService)
	authService := service.NewAuthService(authRepository)
	authHandler := handler.NewAuthHandler(authService, false)
	withSession := middleware.Session(authService)
	progressService := service.NewProgressService(progressRepository, contentRepository)
	progressHandler := handler.NewProgressHandler(progressService)

	mux := http.NewServeMux()
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
	return mux
}

func sessionCookieValue(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			return c.Value
		}
	}
	t.Fatalf("expected a session cookie, got %v", w.Header().Values("Set-Cookie"))
	return ""
}

func loginAs(t *testing.T, h http.Handler, email, password string) (sessionID, csrfToken string) {
	t.Helper()
	w := authRequest(t, h, http.MethodPost, "/auth/login", requestOptions{
		body: `{"email":"` + email + `","password":"` + password + `"}`,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	sessionID = sessionCookieValue(t, w)
	body := decodeBody(t, w)
	csrf, _ := body["csrf_token"].(string)
	if csrf == "" {
		t.Fatalf("login response missing csrf_token: %v", body)
	}
	return sessionID, csrf
}

func quizContainsQuestion(t *testing.T, h http.Handler, topicSlug, questionID string) bool {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/topics/"+topicSlug+"/quiz", strings.NewReader(""))
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start quiz: %d %s", w.Code, w.Body.String())
	}
	var qs []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &qs); err != nil {
		t.Fatalf("decode quiz: %v", err)
	}
	for _, q := range qs {
		if q["id"] == questionID {
			return true
		}
	}
	return false
}

func TestMaintainerSeedAndDraftPublishFlow_EndToEnd(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	resetIntegrationTables(t, pool)
	applyIntegrationMigrations(t, pool)

	const maintainerEmail = "maintainer-seed@example.com"
	const maintainerPassword = "seed-password-123"
	if _, err := seed.Load(ctx, pool, seeddata.Files, maintainerEmail, maintainerPassword); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h := buildIntegrationHandler(t, pool)
	sessionID, csrfToken := loginAs(t, h, maintainerEmail, maintainerPassword)

	// Create draft lesson and question.
	w := authRequest(t, h, http.MethodPost, "/topics/llms/lessons", requestOptions{
		body:   `{"title":"Seed flow lesson","body":"Body","position":99}`,
		cookie: sessionID, csrf: csrfToken,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create draft lesson: %d %s", w.Code, w.Body.String())
	}
	lessonID, _ := decodeBody(t, w)["id"].(string)
	if lessonID == "" {
		t.Fatalf("draft lesson missing id: %s", w.Body.String())
	}

	w = authRequest(t, h, http.MethodPost, "/topics/llms/questions", requestOptions{
		body:   `{"text":"Seed flow question?","options":[{"text":"Yes","is_correct":true},{"text":"No","is_correct":false}]}`,
		cookie: sessionID, csrf: csrfToken,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create draft question: %d %s", w.Code, w.Body.String())
	}
	questionID, _ := decodeBody(t, w)["id"].(string)
	if questionID == "" {
		t.Fatalf("draft question missing id: %s", w.Body.String())
	}

	// Neither visible through public endpoints.
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/lessons/"+lessonID, nil)
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("draft lesson should be hidden before publish, got %d: %s", w.Code, w.Body.String())
	}
	if quizContainsQuestion(t, h, "llms", questionID) {
		t.Fatal("draft question should be hidden from quiz before publish")
	}

	// Publish both.
	w = authRequest(t, h, http.MethodPut, "/topics/llms/lessons/"+lessonID+"/publish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("publish lesson: %d %s", w.Code, w.Body.String())
	}
	w = authRequest(t, h, http.MethodPut, "/topics/llms/questions/"+questionID+"/publish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("publish question: %d %s", w.Code, w.Body.String())
	}

	// Both visible.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/lessons/"+lessonID, nil)
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("published lesson should be visible, got %d: %s", w.Code, w.Body.String())
	}
	if !quizContainsQuestion(t, h, "llms", questionID) {
		t.Fatal("published question should appear in quiz")
	}

	// Unpublish both.
	w = authRequest(t, h, http.MethodPut, "/topics/llms/lessons/"+lessonID+"/unpublish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("unpublish lesson: %d %s", w.Code, w.Body.String())
	}
	w = authRequest(t, h, http.MethodPut, "/topics/llms/questions/"+questionID+"/unpublish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("unpublish question: %d %s", w.Code, w.Body.String())
	}

	// Neither visible again.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/lessons/"+lessonID, nil)
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unpublished lesson should be hidden, got %d: %s", w.Code, w.Body.String())
	}
	if quizContainsQuestion(t, h, "llms", questionID) {
		t.Fatal("unpublished question should be hidden from quiz")
	}
}

func TestDraftPublishOperations_RealDB_RejectNormalUser(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	resetIntegrationTables(t, pool)
	applyIntegrationMigrations(t, pool)
	if _, err := seed.Load(ctx, pool, seeddata.Files, "", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h := buildIntegrationHandler(t, pool)
	// Register a normal (non-maintainer) user.
	w := authRequest(t, h, http.MethodPost, "/auth/register", requestOptions{
		body: `{"email":"normal@example.com","password":"normal-password"}`,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	sessionID := sessionCookieValue(t, w)
	csrf, _ := decodeBody(t, w)["csrf_token"].(string)

	ops := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create lesson", http.MethodPost, "/topics/llms/lessons", `{"title":"T","body":"B","position":1}`},
		{"update lesson", http.MethodPut, "/topics/llms/lessons/any-id", `{"title":"T","body":"B","position":1}`},
		{"create question", http.MethodPost, "/topics/llms/questions", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
		{"update question", http.MethodPut, "/topics/llms/questions/any-id", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
		{"publish lesson", http.MethodPut, "/topics/llms/lessons/any-id/publish", ""},
		{"unpublish lesson", http.MethodPut, "/topics/llms/lessons/any-id/unpublish", ""},
		{"publish question", http.MethodPut, "/topics/llms/questions/any-id/publish", ""},
		{"unpublish question", http.MethodPut, "/topics/llms/questions/any-id/unpublish", ""},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			w := authRequest(t, h, op.method, op.path, requestOptions{body: op.body, cookie: sessionID, csrf: csrf})
			assertErrorResponse(t, w, http.StatusForbidden, "authorization_error")
		})
	}
}

func TestDraftPublishOperations_RealDB_RejectNoSession(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	resetIntegrationTables(t, pool)
	applyIntegrationMigrations(t, pool)
	if _, err := seed.Load(ctx, pool, seeddata.Files, "", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h := buildIntegrationHandler(t, pool)
	ops := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create lesson", http.MethodPost, "/topics/llms/lessons", `{"title":"T","body":"B","position":1}`},
		{"update lesson", http.MethodPut, "/topics/llms/lessons/any-id", `{"title":"T","body":"B","position":1}`},
		{"create question", http.MethodPost, "/topics/llms/questions", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
		{"update question", http.MethodPut, "/topics/llms/questions/any-id", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
		{"publish lesson", http.MethodPut, "/topics/llms/lessons/any-id/publish", ""},
		{"unpublish lesson", http.MethodPut, "/topics/llms/lessons/any-id/unpublish", ""},
		{"publish question", http.MethodPut, "/topics/llms/questions/any-id/publish", ""},
		{"unpublish question", http.MethodPut, "/topics/llms/questions/any-id/unpublish", ""},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			w := authRequest(t, h, op.method, op.path, requestOptions{body: op.body})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		})
	}
}
