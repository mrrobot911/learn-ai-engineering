package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

func post(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, nil))
	return w
}

func TestStartQuizReturnsFiveQuestionsWithoutAnswers(t *testing.T) {
	w := post(t, newFixtureServer(), "/topics/intro/quiz")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	assertContentTypeJSON(t, w)

	body := decodeArray(t, w)
	if len(body) != 5 {
		t.Fatalf("expected 5 questions, got %d: %v", len(body), body)
	}

	published := map[string]bool{
		"question-1": true,
		"question-2": true,
		"question-3": true,
		"question-4": true,
		"question-5": true,
		"question-6": true,
	}
	seen := make(map[string]bool, len(body))
	for _, question := range body {
		assertExactKeys(t, question, []string{"id", "text", "options"})

		id, ok := question["id"].(string)
		if !ok || !published[id] {
			t.Errorf("expected a published question of the topic, got %v", question["id"])
		}
		if seen[id] {
			t.Errorf("question %q returned twice in one quiz", id)
		}
		seen[id] = true
		if text, ok := question["text"].(string); !ok || text == "" {
			t.Errorf("expected a non-empty question text, got %v", question["text"])
		}

		options, ok := question["options"].([]any)
		if !ok || len(options) == 0 {
			t.Fatalf("expected a non-empty options array, got %v", question["options"])
		}
		for _, raw := range options {
			option, ok := raw.(map[string]any)
			if !ok {
				t.Fatalf("expected an option object, got %v", raw)
			}
			assertExactKeys(t, option, []string{"id", "text"})
			if text, ok := option["text"].(string); !ok || text == "" {
				t.Errorf("expected a non-empty option text, got %v", option["text"])
			}
		}
	}

	for _, leak := range []string{"is_correct", "explanation", "correct_option_id"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("quiz response reveals %q: %s", leak, w.Body.String())
		}
	}
}

func TestStartQuizReturnsAllQuestionsWhenFewerThanFive(t *testing.T) {
	w := post(t, newFixtureServer(), "/topics/advanced/quiz")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeArray(t, w)
	if len(body) != 1 {
		t.Fatalf("expected the single published question, got %d: %v", len(body), body)
	}
	if body[0]["id"] != "question-advanced" {
		t.Errorf("expected question-advanced, got %v", body[0]["id"])
	}
}

func TestStartQuizReturnsEmptyArray(t *testing.T) {
	h := newContentTestServer(repo.NewMemoryContentRepository([]repo.Topic{
		{ID: "topic-1", Slug: "intro", Position: 1, Status: repo.StatusPublished},
	}, nil, nil))

	w := post(t, h, "/topics/intro/quiz")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if body := strings.TrimSpace(w.Body.String()); body != "[]" {
		t.Errorf("expected the JSON array [], got %q", body)
	}
}

func TestStartQuizNotFound(t *testing.T) {
	for _, tt := range []struct {
		name string
		slug string
	}{
		{"unknown slug", "unknown"},
		{"draft topic", "draft-topic"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := post(t, newFixtureServer(), "/topics/"+tt.slug+"/quiz")
			assertErrorCode(t, w, http.StatusNotFound, "not_found")
		})
	}
}

func TestStartQuizWorksWithoutSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/topics/intro/quiz", nil)
	if _, err := req.Cookie("session"); err == nil {
		t.Fatal("expected a request without a session cookie")
	}
	w := httptest.NewRecorder()
	newFixtureServer().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 without a session, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStartQuizHidesRepositoryErrors(t *testing.T) {
	cause := errors.New("query failed for user 42: password_hash=secret")
	h := newContentTestServer(failingRepository{err: cause})

	w := post(t, h, "/topics/intro/quiz")

	assertErrorCode(t, w, http.StatusInternalServerError, "internal_error")
	for _, leak := range []string{"query failed", "user 42", "password_hash", "secret"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("error body leaks %q: %s", leak, w.Body.String())
		}
	}
}
