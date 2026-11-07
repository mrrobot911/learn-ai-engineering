package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

type draftServer struct {
	handler     http.Handler
	authRepo    *repo.MemoryAuthRepository
	contentRepo *repo.MemoryContentRepository
}

func newDraftTestServer(t *testing.T) *draftServer {
	t.Helper()
	contentRepo := repo.NewMemoryContentRepository(
		[]repo.Topic{
			{ID: "topic-1", Slug: "intro", Title: "Introduction", Position: 1, Status: repo.StatusPublished},
		},
		[]repo.Lesson{
			{ID: "lesson-draft-1", TopicID: "topic-1", Title: "Draft lesson", Body: "Body", Position: 1, Status: repo.StatusDraft},
		},
		[]repo.Question{
			{
				ID:      "question-draft-1",
				TopicID: "topic-1",
				Text:    "Draft question?",
				Status:  repo.StatusDraft,
				Options: []repo.Option{
					{ID: "o1", Text: "A", Position: 1, IsCorrect: true},
					{ID: "o2", Text: "B", Position: 2},
				},
			},
		},
	)
	authRepo := repo.NewMemoryAuthRepository()
	authService := service.NewAuthService(authRepo)
	contentService := service.NewContentService(contentRepo, repo.NewMemoryAttemptRepository())
	contentHandler := handler.NewContentHandler(contentService)

	withSession := middleware.Session(authService)
	mux := http.NewServeMux()
	mux.Handle("POST /topics/{topicSlug}/lessons", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.CreateDraftLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UpdateDraftLesson)))))
	mux.Handle("POST /topics/{topicSlug}/questions", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.CreateDraftQuestion)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UpdateDraftQuestion)))))

	return &draftServer{handler: mux, authRepo: authRepo, contentRepo: contentRepo}
}

func (s *draftServer) createSession(t *testing.T, userID string, isMaintainer bool) (sessionID, csrfToken string) {
	t.Helper()
	if isMaintainer {
		userID = "maintainer-" + userID
	} else {
		userID = "user-" + userID
	}
	sessionID = "session-" + userID
	csrfToken = "csrf-" + userID
	now := time.Now()
	if err := s.authRepo.CreateUser(context.Background(), repo.User{
		ID:           userID,
		Email:        userID + "@example.com",
		PasswordHash: "hash",
		IsMaintainer: isMaintainer,
		CreatedAt:    now,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := s.authRepo.CreateSession(context.Background(), repo.Session{
		ID:        sessionID,
		UserID:    userID,
		CSRFToken: csrfToken,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return sessionID, csrfToken
}

func draftRequest(t *testing.T, h http.Handler, method, path string, opts requestOptions) *httptest.ResponseRecorder {
	return authRequest(t, h, method, path, opts)
}

func draftOperations() []struct {
	name   string
	method string
	path   string
	body   string
} {
	return []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create lesson", http.MethodPost, "/topics/intro/lessons", `{"title":"T","body":"B","position":1}`},
		{"update lesson", http.MethodPut, "/topics/intro/lessons/lesson-draft-1", `{"title":"T","body":"B","position":1}`},
		{"create question", http.MethodPost, "/topics/intro/questions", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
		{"update question", http.MethodPut, "/topics/intro/questions/question-draft-1", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`},
	}
}

func TestDraftOperationsRejectMissingSession(t *testing.T) {
	for _, op := range draftOperations() {
		t.Run(op.name, func(t *testing.T) {
			s := newDraftTestServer(t)
			w := draftRequest(t, s.handler, op.method, op.path, requestOptions{body: op.body})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		})
	}
}

func TestDraftOperationsRejectNonMaintainer(t *testing.T) {
	for _, op := range draftOperations() {
		t.Run(op.name, func(t *testing.T) {
			s := newDraftTestServer(t)
			sessionID, csrfToken := s.createSession(t, "regular", false)
			w := draftRequest(t, s.handler, op.method, op.path, requestOptions{body: op.body, cookie: sessionID, csrf: csrfToken})
			assertErrorResponse(t, w, http.StatusForbidden, "authorization_error")
		})
	}
}

func TestDraftOperationsRejectSessionWithoutMaintainer(t *testing.T) {
	// Explicit duplicate of the acceptance-criteria case naming.
	for _, op := range draftOperations() {
		t.Run(op.name, func(t *testing.T) {
			s := newDraftTestServer(t)
			sessionID, csrfToken := s.createSession(t, "nomaint", false)
			w := draftRequest(t, s.handler, op.method, op.path, requestOptions{body: op.body, cookie: sessionID, csrf: csrfToken})
			assertErrorResponse(t, w, http.StatusForbidden, "authorization_error")
		})
	}
}

func TestCreateDraftLessonStoresAsDraft(t *testing.T) {
	s := newDraftTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m1", true)
	w := draftRequest(t, s.handler, http.MethodPost, "/topics/intro/lessons", requestOptions{
		body:   `{"title":"New","body":"Content","position":5}`,
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	assertJSON(t, w)
	body := decodeBody(t, w)
	if body["status"] != "draft" {
		t.Errorf("expected status draft, got %v", body["status"])
	}
	if body["topic_id"] != "topic-1" {
		t.Errorf("expected topic_id topic-1, got %v", body["topic_id"])
	}
	assertExactKeys(t, body, "id", "topic_id", "title", "body", "position", "status")
}

func TestUpdateDraftLessonStoresAsDraft(t *testing.T) {
	s := newDraftTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m2", true)
	w := draftRequest(t, s.handler, http.MethodPut, "/topics/intro/lessons/lesson-draft-1", requestOptions{
		body:   `{"title":"Updated","body":"New body","position":2}`,
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["status"] != "draft" {
		t.Errorf("expected status draft, got %v", body["status"])
	}
	if body["title"] != "Updated" {
		t.Errorf("expected title Updated, got %v", body["title"])
	}
}

func TestCreateDraftQuestionStoresAsDraft(t *testing.T) {
	s := newDraftTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m3", true)
	w := draftRequest(t, s.handler, http.MethodPost, "/topics/intro/questions", requestOptions{
		body:   `{"text":"What?","options":[{"text":"Yes","is_correct":true},{"text":"No","is_correct":false}]}`,
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["status"] != "draft" {
		t.Errorf("expected status draft, got %v", body["status"])
	}
	assertExactKeys(t, body, "id", "topic_id", "text", "options", "status")
	options, ok := body["options"].([]any)
	if !ok || len(options) != 2 {
		t.Fatalf("expected 2 options, got %v", body["options"])
	}
	opt0 := options[0].(map[string]any)
	assertExactKeys(t, opt0, "id", "text", "is_correct")
}

func TestUpdateDraftQuestionStoresAsDraft(t *testing.T) {
	s := newDraftTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m4", true)
	w := draftRequest(t, s.handler, http.MethodPut, "/topics/intro/questions/question-draft-1", requestOptions{
		body:   `{"text":"Updated?","options":[{"text":"Yes","is_correct":false},{"text":"No","is_correct":true}]}`,
		cookie: sessionID,
		csrf:   csrfToken,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["status"] != "draft" {
		t.Errorf("expected status draft, got %v", body["status"])
	}
}

func TestDraftQuestionValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"too few options", `{"text":"Q?","options":[{"text":"A","is_correct":true}]}`},
		{"too many options", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false},{"text":"C","is_correct":false},{"text":"D","is_correct":false},{"text":"E","is_correct":false},{"text":"F","is_correct":false},{"text":"G","is_correct":false}]}`},
		{"no correct option", `{"text":"Q?","options":[{"text":"A","is_correct":false},{"text":"B","is_correct":false}]}`},
		{"multiple correct options", `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":true}]}`},
	}
	for _, tt := range tests {
		t.Run("create "+tt.name, func(t *testing.T) {
			s := newDraftTestServer(t)
			sessionID, csrfToken := s.createSession(t, "v1-"+tt.name, true)
			w := draftRequest(t, s.handler, http.MethodPost, "/topics/intro/questions", requestOptions{
				body: tt.body, cookie: sessionID, csrf: csrfToken,
			})
			assertErrorResponse(t, w, http.StatusBadRequest, "validation_error")
		})
		t.Run("update "+tt.name, func(t *testing.T) {
			s := newDraftTestServer(t)
			sessionID, csrfToken := s.createSession(t, "v2-"+tt.name, true)
			w := draftRequest(t, s.handler, http.MethodPut, "/topics/intro/questions/question-draft-1", requestOptions{
				body: tt.body, cookie: sessionID, csrf: csrfToken,
			})
			assertErrorResponse(t, w, http.StatusBadRequest, "validation_error")
		})
	}
}

func TestCreateDraftQuestionRejectsUnknownTopic(t *testing.T) {
	s := newDraftTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m404", true)
	w := draftRequest(t, s.handler, http.MethodPost, "/topics/nope/questions", requestOptions{
		body:   `{"text":"Q?","options":[{"text":"A","is_correct":true},{"text":"B","is_correct":false}]}`,
		cookie: sessionID,
		csrf:   csrfToken,
	})
	assertErrorResponse(t, w, http.StatusNotFound, "not_found")
}

var _ = json.Unmarshal
var _ = httptest.NewRequest
var _ = strings.NewReader
