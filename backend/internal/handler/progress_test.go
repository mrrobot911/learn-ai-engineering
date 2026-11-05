package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

type progressClock struct {
	now time.Time
}

func (c *progressClock) Now() time.Time {
	return c.now
}

func (c *progressClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

type progressServer struct {
	*authServer
	progressRepo *repo.MemoryProgressRepository
	clock        *progressClock
}

func buildProgressHandler(
	authRepo repo.AuthRepository,
	contentRepo repo.ContentRepository,
	progressRepo repo.ProgressRepository,
	clock *progressClock,
) http.Handler {
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService, false)
	progressHandler := handler.NewProgressHandler(service.NewProgressServiceWithClock(progressRepo, contentRepo, clock.Now))

	withSession := middleware.Session(authService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.Handle("GET /me/progress", withSession(http.HandlerFunc(progressHandler.GetMyProgress)))
	mux.Handle("PUT /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.MarkLessonComplete))))
	mux.Handle("DELETE /me/progress/lessons/{lessonId}", withSession(middleware.CSRF(http.HandlerFunc(progressHandler.UnmarkLessonComplete))))
	return mux
}

func newProgressTestServer(t *testing.T) *progressServer {
	t.Helper()

	contentRepo := repo.NewMemoryContentRepository(
		[]repo.Topic{
			{ID: "topic-1", Slug: "intro", Title: "Introduction", Position: 1, Status: repo.StatusPublished},
		},
		[]repo.Lesson{
			{ID: "lesson-1", TopicID: "topic-1", Title: "First lesson", Position: 1, Status: repo.StatusPublished},
			{ID: "lesson-2", TopicID: "topic-1", Title: "Second lesson", Position: 2, Status: repo.StatusPublished},
			{ID: "lesson-draft", TopicID: "topic-1", Title: "Draft lesson", Position: 3, Status: repo.StatusDraft},
		},
		nil,
	)
	authRepo := repo.NewMemoryAuthRepository()
	progressRepo := repo.NewMemoryProgressRepository()
	clock := &progressClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}

	return &progressServer{
		authServer: &authServer{
			handler:  buildProgressHandler(authRepo, contentRepo, progressRepo, clock),
			authRepo: authRepo,
		},
		progressRepo: progressRepo,
		clock:        clock,
	}
}

func (s *progressServer) register(t *testing.T, email string) (userID, sessionID, csrfToken string) {
	t.Helper()
	sessionID, csrfToken, _ = registerUser(t, s.authServer, email, "password")
	user, err := s.authRepo.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("get registered user: %v", err)
	}
	return user.ID, sessionID, csrfToken
}

func progressIDs(t *testing.T, body []byte) []string {
	t.Helper()
	var ids []string
	if err := json.Unmarshal(body, &ids); err != nil {
		t.Fatalf("decode progress response: %v; body: %s", err, body)
	}
	if ids == nil {
		t.Fatalf("expected a JSON array, got null")
	}
	return ids
}

func assertProgressIDs(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected lesson ids %v, got %v", want, got)
	}
	seen := make(map[string]bool, len(got))
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			t.Errorf("expected lesson id %q in %v", id, got)
		}
	}
}

func assertProgressRecord(t *testing.T, s *progressServer, userID, lessonID string) time.Time {
	t.Helper()
	records, err := s.progressRepo.ListLessonProgress(context.Background(), userID)
	if err != nil {
		t.Fatalf("list stored progress: %v", err)
	}
	for _, record := range records {
		if record.LessonID == lessonID {
			if record.CompletedAt.IsZero() {
				t.Fatalf("expected a non-zero completion time for lesson %q", lessonID)
			}
			return record.CompletedAt
		}
	}
	t.Fatalf("expected a stored progress record for lesson %q, got %v", lessonID, records)
	return time.Time{}
}

func assertNoProgressRecord(t *testing.T, s *progressServer, userID, lessonID string) {
	t.Helper()
	records, err := s.progressRepo.ListLessonProgress(context.Background(), userID)
	if err != nil {
		t.Fatalf("list stored progress: %v", err)
	}
	for _, record := range records {
		if record.LessonID == lessonID {
			t.Fatalf("expected no progress record for lesson %q, got %v", lessonID, record)
		}
	}
}

func TestMarkLessonComplete(t *testing.T) {
	s := newProgressTestServer(t)
	userID, sessionID, csrfToken := s.register(t, "user@example.com")

	w := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/lesson-1", requestOptions{
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	assertJSON(t, w)
	assertProgressIDs(t, progressIDs(t, w.Body.Bytes()), "lesson-1")
	assertProgressRecord(t, s, userID, "lesson-1")
}

func TestMarkLessonCompleteIsIdempotent(t *testing.T) {
	s := newProgressTestServer(t)
	userID, sessionID, csrfToken := s.register(t, "user@example.com")

	first := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/lesson-1", requestOptions{
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if first.Code != http.StatusOK {
		t.Fatalf("first mark: expected status 200, got %d: %s", first.Code, first.Body.String())
	}
	completedAt := assertProgressRecord(t, s, userID, "lesson-1")

	s.clock.Advance(time.Hour)

	second := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/lesson-1", requestOptions{
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if second.Code != http.StatusOK {
		t.Fatalf("second mark: expected status 200, got %d: %s", second.Code, second.Body.String())
	}
	assertProgressIDs(t, progressIDs(t, second.Body.Bytes()), "lesson-1")
	if got := assertProgressRecord(t, s, userID, "lesson-1"); !got.Equal(completedAt) {
		t.Errorf("expected the completion time %s to stay unchanged, got %s", completedAt, got)
	}
}

func TestGetMyProgress(t *testing.T) {
	t.Run("empty progress is an empty array", func(t *testing.T) {
		s := newProgressTestServer(t)
		_, sessionID, _ := s.register(t, "user@example.com")

		w := authRequest(t, s.handler, http.MethodGet, "/me/progress", requestOptions{cookie: sessionID})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "[]" {
			t.Errorf("expected the body [], got %q", got)
		}
		if ids := progressIDs(t, w.Body.Bytes()); len(ids) != 0 {
			t.Errorf("expected no lesson ids, got %v", ids)
		}
	})

	t.Run("returns only the session user's completed lessons", func(t *testing.T) {
		s := newProgressTestServer(t)
		_, sessionID, csrfToken := s.register(t, "user@example.com")

		for _, lessonID := range []string{"lesson-1", "lesson-2"} {
			w := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/"+lessonID, requestOptions{
				cookie: sessionID,
				csrf:   csrfToken,
			})
			if w.Code != http.StatusOK {
				t.Fatalf("mark %s: expected status 200, got %d: %s", lessonID, w.Code, w.Body.String())
			}
		}

		w := authRequest(t, s.handler, http.MethodGet, "/me/progress", requestOptions{cookie: sessionID})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		assertProgressIDs(t, progressIDs(t, w.Body.Bytes()), "lesson-1", "lesson-2")
	})
}

func TestUnmarkLessonComplete(t *testing.T) {
	t.Run("removes a marked lesson", func(t *testing.T) {
		s := newProgressTestServer(t)
		userID, sessionID, csrfToken := s.register(t, "user@example.com")

		mark := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/lesson-1", requestOptions{
			cookie: sessionID,
			csrf:   csrfToken,
		})
		if mark.Code != http.StatusOK {
			t.Fatalf("mark: expected status 200, got %d: %s", mark.Code, mark.Body.String())
		}

		w := authRequest(t, s.handler, http.MethodDelete, "/me/progress/lessons/lesson-1", requestOptions{
			cookie: sessionID,
			csrf:   csrfToken,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		assertJSON(t, w)
		assertProgressIDs(t, progressIDs(t, w.Body.Bytes()))
		assertNoProgressRecord(t, s, userID, "lesson-1")
	})

	t.Run("undoing an unmarked lesson is not an error", func(t *testing.T) {
		s := newProgressTestServer(t)
		_, sessionID, csrfToken := s.register(t, "user@example.com")

		w := authRequest(t, s.handler, http.MethodDelete, "/me/progress/lessons/lesson-1", requestOptions{
			cookie: sessionID,
			csrf:   csrfToken,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		assertProgressIDs(t, progressIDs(t, w.Body.Bytes()))
	})
}

func TestProgressEndpointsRequireASession(t *testing.T) {
	s := newProgressTestServer(t)
	userID, _, _ := s.register(t, "user@example.com")

	expired := repo.Session{
		ID:        "expired-session",
		UserID:    userID,
		CSRFToken: "expired-token",
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if err := s.authRepo.CreateSession(context.Background(), expired); err != nil {
		t.Fatalf("create expired session: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		cookie string
	}{
		{"get without a cookie", http.MethodGet, "/me/progress", ""},
		{"put without a cookie", http.MethodPut, "/me/progress/lessons/lesson-1", ""},
		{"delete without a cookie", http.MethodDelete, "/me/progress/lessons/lesson-1", ""},
		{"get with an invalid cookie", http.MethodGet, "/me/progress", "invalid-session"},
		{"put with an invalid cookie", http.MethodPut, "/me/progress/lessons/lesson-1", "invalid-session"},
		{"delete with an invalid cookie", http.MethodDelete, "/me/progress/lessons/lesson-1", "invalid-session"},
		{"get with an expired cookie", http.MethodGet, "/me/progress", expired.ID},
		{"put with an expired cookie", http.MethodPut, "/me/progress/lessons/lesson-1", expired.ID},
		{"delete with an expired cookie", http.MethodDelete, "/me/progress/lessons/lesson-1", expired.ID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := authRequest(t, s.handler, tt.method, tt.path, requestOptions{cookie: tt.cookie})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		})
	}
}

func TestProgressWritesRequireCSRF(t *testing.T) {
	for _, tt := range []struct {
		name  string
		token string
	}{
		{"missing token", ""},
		{"wrong token", "wrong-token"},
	} {
		t.Run("put with "+tt.name, func(t *testing.T) {
			s := newProgressTestServer(t)
			userID, sessionID, _ := s.register(t, "user@example.com")

			w := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/lesson-1", requestOptions{
				cookie: sessionID,
				csrf:   tt.token,
			})
			assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
			assertNoProgressRecord(t, s, userID, "lesson-1")
		})

		t.Run("delete with "+tt.name, func(t *testing.T) {
			s := newProgressTestServer(t)
			userID, sessionID, _ := s.register(t, "user@example.com")
			if err := s.progressRepo.MarkLessonComplete(context.Background(), userID, "lesson-1", s.clock.Now()); err != nil {
				t.Fatalf("seed progress: %v", err)
			}

			w := authRequest(t, s.handler, http.MethodDelete, "/me/progress/lessons/lesson-1", requestOptions{
				cookie: sessionID,
				csrf:   tt.token,
			})
			assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
			assertProgressRecord(t, s, userID, "lesson-1")
		})
	}
}

func TestProgressWritesRejectUnknownAndDraftLessons(t *testing.T) {
	for _, lessonID := range []string{"lesson-unknown", "lesson-draft"} {
		t.Run("put "+lessonID, func(t *testing.T) {
			s := newProgressTestServer(t)
			userID, sessionID, csrfToken := s.register(t, "user@example.com")

			w := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/"+lessonID, requestOptions{
				cookie: sessionID,
				csrf:   csrfToken,
			})
			assertErrorResponse(t, w, http.StatusNotFound, "not_found")
			assertNoProgressRecord(t, s, userID, lessonID)
		})

		t.Run("delete "+lessonID, func(t *testing.T) {
			s := newProgressTestServer(t)
			userID, sessionID, csrfToken := s.register(t, "user@example.com")
			if err := s.progressRepo.MarkLessonComplete(context.Background(), userID, lessonID, s.clock.Now()); err != nil {
				t.Fatalf("seed progress: %v", err)
			}

			w := authRequest(t, s.handler, http.MethodDelete, "/me/progress/lessons/"+lessonID, requestOptions{
				cookie: sessionID,
				csrf:   csrfToken,
			})
			assertErrorResponse(t, w, http.StatusNotFound, "not_found")
			assertProgressRecord(t, s, userID, lessonID)
		})
	}
}

func TestProgressIsIsolatedPerUser(t *testing.T) {
	s := newProgressTestServer(t)
	firstUserID, firstSessionID, firstToken := s.register(t, "first@example.com")
	_, secondSessionID, secondToken := s.register(t, "second@example.com")

	mark := func(sessionID, csrfToken, lessonID string) {
		t.Helper()
		w := authRequest(t, s.handler, http.MethodPut, "/me/progress/lessons/"+lessonID, requestOptions{
			cookie: sessionID,
			csrf:   csrfToken,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("mark %s: expected status 200, got %d: %s", lessonID, w.Code, w.Body.String())
		}
	}
	mark(firstSessionID, firstToken, "lesson-1")
	mark(secondSessionID, secondToken, "lesson-2")

	first := authRequest(t, s.handler, http.MethodGet, "/me/progress?user_id="+firstUserID, requestOptions{cookie: firstSessionID})
	if first.Code != http.StatusOK {
		t.Fatalf("first user progress: expected status 200, got %d: %s", first.Code, first.Body.String())
	}
	assertProgressIDs(t, progressIDs(t, first.Body.Bytes()), "lesson-1")

	second := authRequest(t, s.handler, http.MethodGet, "/me/progress", requestOptions{cookie: secondSessionID})
	if second.Code != http.StatusOK {
		t.Fatalf("second user progress: expected status 200, got %d: %s", second.Code, second.Body.String())
	}
	assertProgressIDs(t, progressIDs(t, second.Body.Bytes()), "lesson-2")

	body := `{"user_id":"` + firstUserID + `"}`
	other := authRequest(t, s.handler, http.MethodGet, "/me/progress", requestOptions{cookie: secondSessionID, body: body})
	if other.Code != http.StatusOK {
		t.Fatalf("progress with a body user id: expected status 200, got %d: %s", other.Code, other.Body.String())
	}
	assertProgressIDs(t, progressIDs(t, other.Body.Bytes()), "lesson-2")
}
