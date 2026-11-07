package handler_test

import (
	"context"
	"encoding/json"
	"math/rand"
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

type publishServer struct {
	handler     http.Handler
	public      http.Handler
	authRepo    *repo.MemoryAuthRepository
	contentRepo *repo.MemoryContentRepository
}

func newPublishTestServer(t *testing.T) *publishServer {
	t.Helper()

	contentRepo := repo.NewMemoryContentRepository(
		[]repo.Topic{
			{ID: "topic-1", Slug: "intro", Title: "Introduction", Position: 1, Status: repo.StatusPublished},
		},
		[]repo.Lesson{
			{ID: "lesson-draft-1", TopicID: "topic-1", Title: "Draft lesson", Body: "Body", Position: 1, Status: repo.StatusDraft},
			{ID: "lesson-pub-1", TopicID: "topic-1", Title: "Published lesson", Body: "Pub body", Position: 2, Status: repo.StatusPublished},
		},
		[]repo.Question{
			{
				ID:      "question-draft-1",
				TopicID: "topic-1",
				Text:    "Draft question?",
				Status:  repo.StatusDraft,
				Options: []repo.Option{{ID: "do1", Text: "A", Position: 1, IsCorrect: true}, {ID: "do2", Text: "B", Position: 2}},
			},
			{
				ID:      "question-pub-1",
				TopicID: "topic-1",
				Text:    "Published question?",
				Status:  repo.StatusPublished,
				Options: []repo.Option{{ID: "po1", Text: "A", Position: 1, IsCorrect: true}, {ID: "po2", Text: "B", Position: 2}},
			},
			{
				ID:      "question-pub-2",
				TopicID: "topic-1",
				Text:    "Published question 2?",
				Status:  repo.StatusPublished,
				Options: []repo.Option{{ID: "p21", Text: "A", Position: 1, IsCorrect: true}, {ID: "p22", Text: "B", Position: 2}},
			},
			{
				ID:      "question-pub-3",
				TopicID: "topic-1",
				Text:    "Published question 3?",
				Status:  repo.StatusPublished,
				Options: []repo.Option{{ID: "p31", Text: "A", Position: 1, IsCorrect: true}, {ID: "p32", Text: "B", Position: 2}},
			},
			{
				ID:      "question-pub-4",
				TopicID: "topic-1",
				Text:    "Published question 4?",
				Status:  repo.StatusPublished,
				Options: []repo.Option{{ID: "p41", Text: "A", Position: 1, IsCorrect: true}, {ID: "p42", Text: "B", Position: 2}},
			},
		},
	)
	authRepo := repo.NewMemoryAuthRepository()
	authService := service.NewAuthService(authRepo)
	withSession := middleware.Session(authService)

	contentService := service.NewContentServiceWithRandom(contentRepo, repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(1)))
	contentHandler := handler.NewContentHandler(contentService)

	mux := http.NewServeMux()
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}/publish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.PublishLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/lessons/{lessonId}/unpublish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UnpublishLesson)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}/publish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.PublishQuestion)))))
	mux.Handle("PUT /topics/{topicSlug}/questions/{questionId}/unpublish", withSession(middleware.CSRF(middleware.Maintainer(http.HandlerFunc(contentHandler.UnpublishQuestion)))))

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("GET /topics/{topicSlug}", contentHandler.GetTopic)
	publicMux.HandleFunc("GET /lessons/{lessonId}", contentHandler.GetLesson)
	publicMux.HandleFunc("POST /topics/{topicSlug}/quiz", contentHandler.StartQuiz)

	return &publishServer{handler: mux, public: publicMux, authRepo: authRepo, contentRepo: contentRepo}
}

func (s *publishServer) createSession(t *testing.T, userID string, isMaintainer bool) (sessionID, csrfToken string) {
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

func publishOperations() []struct {
	name   string
	method string
	path   string
} {
	return []struct {
		name   string
		method string
		path   string
	}{
		{"publish lesson", http.MethodPut, "/topics/intro/lessons/lesson-draft-1/publish"},
		{"unpublish lesson", http.MethodPut, "/topics/intro/lessons/lesson-pub-1/unpublish"},
		{"publish question", http.MethodPut, "/topics/intro/questions/question-draft-1/publish"},
		{"unpublish question", http.MethodPut, "/topics/intro/questions/question-pub-1/unpublish"},
	}
}

func TestPublishOperationsRejectMissingSession(t *testing.T) {
	for _, op := range publishOperations() {
		t.Run(op.name, func(t *testing.T) {
			s := newPublishTestServer(t)
			w := authRequest(t, s.handler, op.method, op.path, requestOptions{})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		})
	}
}

func TestPublishOperationsRejectNonMaintainer(t *testing.T) {
	for _, op := range publishOperations() {
		t.Run(op.name, func(t *testing.T) {
			s := newPublishTestServer(t)
			sessionID, csrfToken := s.createSession(t, "nonmaint", false)
			w := authRequest(t, s.handler, op.method, op.path, requestOptions{cookie: sessionID, csrf: csrfToken})
			assertErrorResponse(t, w, http.StatusForbidden, "authorization_error")
		})
	}
}

func TestPublishLessonSetsStatusAndMakesVisible(t *testing.T) {
	s := newPublishTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m1", true)

	w := authRequest(t, s.handler, http.MethodPut, "/topics/intro/lessons/lesson-draft-1/publish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "published" || body["id"] != "lesson-draft-1" {
		t.Fatalf("expected published lesson-draft-1, got %v", body)
	}
	if body["body"] == nil || body["topic_id"] == nil {
		t.Fatalf("expected Lesson schema fields, got %v", body)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/lessons/lesson-draft-1", nil)
	s.public.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected lesson visible via GET /lessons, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestUnpublishLessonHidesIt(t *testing.T) {
	s := newPublishTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m1", true)

	w := authRequest(t, s.handler, http.MethodPut, "/topics/intro/lessons/lesson-pub-1/unpublish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "draft" || body["id"] != "lesson-pub-1" {
		t.Fatalf("expected DraftLesson with draft status, got %v", body)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/lessons/lesson-pub-1", nil)
	s.public.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected unpublished lesson hidden, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestPublishQuestionMakesVisibleInQuiz(t *testing.T) {
	s := newPublishTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m1", true)

	w := authRequest(t, s.handler, http.MethodPut, "/topics/intro/questions/question-draft-1/publish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "question-draft-1" || body["text"] == nil {
		t.Fatalf("expected QuizQuestion, got %v", body)
	}
	if _, ok := body["status"]; ok {
		t.Errorf("QuizQuestion response must not include status: %v", body)
	}
	if _, ok := body["topic_id"]; ok {
		t.Errorf("QuizQuestion response must not include topic_id: %v", body)
	}
	opts, _ := body["options"].([]any)
	if len(opts) > 0 {
		opt0, _ := opts[0].(map[string]any)
		if _, ok := opt0["is_correct"]; ok {
			t.Errorf("QuizQuestion options must not include is_correct: %v", opt0)
		}
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/topics/intro/quiz", strings.NewReader(""))
	s.public.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected quiz start ok, got %d: %s", w2.Code, w2.Body.String())
	}
	var qs []map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &qs); err != nil {
		t.Fatalf("decode quiz: %v", err)
	}
	found := false
	for _, q := range qs {
		if q["id"] == "question-draft-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected published question in quiz, got %v", qs)
	}
}

func TestUnpublishQuestionHidesFromQuiz(t *testing.T) {
	s := newPublishTestServer(t)
	sessionID, csrfToken := s.createSession(t, "m1", true)

	w := authRequest(t, s.handler, http.MethodPut, "/topics/intro/questions/question-pub-1/unpublish", requestOptions{cookie: sessionID, csrf: csrfToken})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "draft" || body["id"] != "question-pub-1" {
		t.Fatalf("expected DraftQuestion with draft status, got %v", body)
	}
	if _, ok := body["topic_id"]; !ok {
		t.Fatalf("expected DraftQuestion to include topic_id: %v", body)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/topics/intro/quiz", strings.NewReader(""))
	s.public.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected quiz start ok, got %d: %s", w2.Code, w2.Body.String())
	}
	var qs []map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &qs); err != nil {
		t.Fatalf("decode quiz: %v", err)
	}
	for _, q := range qs {
		if q["id"] == "question-pub-1" {
			t.Fatalf("expected unpublished question hidden from quiz, got %v", qs)
		}
	}
}

func TestPublishOperationsNotFound(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"publish lesson unknown topic", "/topics/nope/lessons/lesson-draft-1/publish"},
		{"publish lesson unknown lesson", "/topics/intro/lessons/nope/publish"},
		{"unpublish question unknown question", "/topics/intro/questions/nope/unpublish"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newPublishTestServer(t)
			sessionID, csrfToken := s.createSession(t, "m1", true)
			w := authRequest(t, s.handler, http.MethodPut, tc.path, requestOptions{cookie: sessionID, csrf: csrfToken})
			assertErrorResponse(t, w, http.StatusNotFound, "not_found")
		})
	}
}
