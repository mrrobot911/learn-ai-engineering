package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

const lessonMarkdown = "# First lesson\n\nRead **this** markdown unchanged.\n"

func contentFixtures() ([]repo.Topic, []repo.Lesson) {
	topics := []repo.Topic{
		{ID: "topic-1", Slug: "intro", Title: "Introduction", Description: "Start here.", Position: 1, Status: repo.StatusPublished},
		{ID: "topic-2", Slug: "draft-topic", Title: "Draft topic", Description: "Not visible.", Position: 2, Status: repo.StatusDraft},
		{ID: "topic-3", Slug: "advanced", Title: "Advanced", Description: "No lessons yet.", Position: 3, Status: repo.StatusPublished},
	}
	lessons := []repo.Lesson{
		{ID: "lesson-2", TopicID: "topic-1", Title: "Second lesson", Body: "Second body", Position: 2, Status: repo.StatusPublished},
		{ID: "lesson-1", TopicID: "topic-1", Title: "First lesson", Body: lessonMarkdown, Position: 1, Status: repo.StatusPublished},
		{ID: "lesson-draft", TopicID: "topic-1", Title: "Draft lesson", Body: "Hidden body", Position: 3, Status: repo.StatusDraft},
	}
	return topics, lessons
}

func newContentTestServer(repository repo.ContentRepository) http.Handler {
	h := NewContentHandler(service.NewContentService(repository))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /topics", h.ListTopics)
	mux.HandleFunc("GET /topics/{topicSlug}", h.GetTopic)
	mux.HandleFunc("GET /lessons/{lessonId}", h.GetLesson)
	return mux
}

func newFixtureServer() http.Handler {
	topics, lessons := contentFixtures()
	return newContentTestServer(repo.NewMemoryContentRepository(topics, lessons))
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func decodeObject(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode object: %v; body: %s", err, w.Body.String())
	}
	return body
}

func decodeArray(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var body []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode array: %v; body: %s", err, w.Body.String())
	}
	return body
}

func assertExactKeys(t *testing.T, object map[string]any, want []string) {
	t.Helper()
	if len(object) != len(want) {
		t.Fatalf("expected exactly %d keys %v, got %v", len(want), want, object)
	}
	for _, key := range want {
		if _, ok := object[key]; !ok {
			t.Errorf("expected key %q, got %v", key, object)
		}
	}
}

func assertContentTypeJSON(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", got)
	}
}

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Errorf("expected status %d, got %d", status, w.Code)
	}
	assertContentTypeJSON(t, w)
	body := decodeObject(t, w)
	assertExactKeys(t, body, []string{"code", "message"})
	if body["code"] != code {
		t.Errorf("expected code %q, got %v", code, body["code"])
	}
	if message, ok := body["message"].(string); !ok || message == "" {
		t.Errorf("expected a non-empty message, got %v", body["message"])
	}
}

func TestListTopics(t *testing.T) {
	w := get(t, newFixtureServer(), "/topics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertContentTypeJSON(t, w)

	body := decodeArray(t, w)
	if len(body) != 2 {
		t.Fatalf("expected 2 published topics, got %d: %v", len(body), body)
	}
	if body[0]["id"] != "topic-1" || body[1]["id"] != "topic-3" {
		t.Errorf("expected published topics in position order, got %v", body)
	}
	if body[0]["slug"] != "intro" || body[0]["title"] != "Introduction" || body[0]["description"] != "Start here." {
		t.Errorf("unexpected first topic fields: %v", body[0])
	}
	if body[0]["position"] != float64(1) {
		t.Errorf("expected position 1, got %v", body[0]["position"])
	}
	for _, topic := range body {
		assertExactKeys(t, topic, []string{"id", "slug", "title", "description", "position", "status"})
		if topic["status"] != string(repo.StatusPublished) {
			t.Errorf("expected status published, got %v", topic["status"])
		}
	}
}

func TestListTopicsReturnsEmptyArray(t *testing.T) {
	h := newContentTestServer(repo.NewMemoryContentRepository([]repo.Topic{
		{ID: "topic-1", Slug: "draft", Position: 1, Status: repo.StatusDraft},
	}, nil))

	w := get(t, h, "/topics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != "[]" {
		t.Errorf("expected the JSON array [], got %q", body)
	}
}

func TestGetTopic(t *testing.T) {
	w := get(t, newFixtureServer(), "/topics/intro")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertContentTypeJSON(t, w)

	body := decodeObject(t, w)
	assertExactKeys(t, body, []string{"id", "slug", "title", "description", "position", "status", "lessons"})
	if body["id"] != "topic-1" || body["slug"] != "intro" || body["status"] != string(repo.StatusPublished) {
		t.Errorf("unexpected topic fields: %v", body)
	}

	lessons, ok := body["lessons"].([]any)
	if !ok {
		t.Fatalf("expected a lessons array, got %v", body["lessons"])
	}
	if len(lessons) != 2 {
		t.Fatalf("expected 2 published lessons, got %d: %v", len(lessons), lessons)
	}
	first, ok := lessons[0].(map[string]any)
	if !ok {
		t.Fatalf("expected a lesson object, got %v", lessons[0])
	}
	second, ok := lessons[1].(map[string]any)
	if !ok {
		t.Fatalf("expected a lesson object, got %v", lessons[1])
	}
	if first["id"] != "lesson-1" || second["id"] != "lesson-2" {
		t.Errorf("expected lessons in position order with the draft excluded, got %v", lessons)
	}
	for _, raw := range lessons {
		lesson, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected a lesson object, got %v", raw)
		}
		assertExactKeys(t, lesson, []string{"id", "topic_id", "title", "body", "position", "status"})
	}
}

func TestGetTopicReturnsEmptyLessonArray(t *testing.T) {
	w := get(t, newFixtureServer(), "/topics/advanced")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := decodeObject(t, w)
	lessons, ok := body["lessons"].([]any)
	if !ok {
		t.Fatalf("expected a lessons array, got %v", body["lessons"])
	}
	if lessons == nil {
		t.Fatal("expected lessons [], got null")
	}
	if len(lessons) != 0 {
		t.Errorf("expected no lessons, got %v", lessons)
	}
}

func TestGetTopicNotFound(t *testing.T) {
	for _, tt := range []struct {
		name string
		slug string
	}{
		{"unknown slug", "unknown"},
		{"draft topic", "draft-topic"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := get(t, newFixtureServer(), "/topics/"+tt.slug)
			assertErrorCode(t, w, http.StatusNotFound, "not_found")
		})
	}
}

func TestGetLesson(t *testing.T) {
	w := get(t, newFixtureServer(), "/lessons/lesson-1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertContentTypeJSON(t, w)

	body := decodeObject(t, w)
	assertExactKeys(t, body, []string{"id", "topic_id", "title", "body", "position", "status"})
	if body["id"] != "lesson-1" || body["topic_id"] != "topic-1" {
		t.Errorf("unexpected lesson identity: %v", body)
	}
	if body["title"] != "First lesson" {
		t.Errorf("expected title First lesson, got %v", body["title"])
	}
	if body["body"] != lessonMarkdown {
		t.Errorf("expected the stored Markdown unchanged, got %q", body["body"])
	}
	if body["position"] != float64(1) {
		t.Errorf("expected position 1, got %v", body["position"])
	}
	if body["status"] != string(repo.StatusPublished) {
		t.Errorf("expected status published, got %v", body["status"])
	}
}

func TestGetLessonNotFound(t *testing.T) {
	for _, tt := range []struct {
		name string
		id   string
	}{
		{"unknown id", "unknown"},
		{"draft lesson", "lesson-draft"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := get(t, newFixtureServer(), "/lessons/"+tt.id)
			assertErrorCode(t, w, http.StatusNotFound, "not_found")
		})
	}
}

type failingRepository struct {
	err error
}

func (r failingRepository) ListPublishedTopics(context.Context) ([]repo.Topic, error) {
	return nil, r.err
}

func (r failingRepository) GetPublishedTopicBySlug(context.Context, string) (repo.Topic, error) {
	return repo.Topic{}, r.err
}

func (r failingRepository) ListPublishedLessonsByTopicID(context.Context, string) ([]repo.Lesson, error) {
	return nil, r.err
}

func (r failingRepository) GetPublishedLessonByID(context.Context, string) (repo.Lesson, error) {
	return repo.Lesson{}, r.err
}

func TestContentEndpointsHideRepositoryErrors(t *testing.T) {
	cause := errors.New("query failed for user 42: password_hash=secret")
	h := newContentTestServer(failingRepository{err: cause})

	for _, path := range []string{"/topics", "/topics/intro", "/lessons/lesson-1"} {
		t.Run(path, func(t *testing.T) {
			w := get(t, h, path)

			assertErrorCode(t, w, http.StatusInternalServerError, "internal_error")
			for _, leak := range []string{"query failed", "user 42", "password_hash", "secret"} {
				if strings.Contains(w.Body.String(), leak) {
					t.Errorf("error body leaks %q: %s", leak, w.Body.String())
				}
			}
		})
	}
}
