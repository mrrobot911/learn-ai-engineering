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

func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(w, req)
	return w
}

func TestSubmitQuizGradesOnTheServer(t *testing.T) {
	body := `{"answers":[
		{"question_id":"question-1","option_id":"question-1-option-1"},
		{"question_id":"question-2","option_id":"question-2-option-2"},
		{"question_id":"question-3","option_id":null},
		{"question_id":"question-4","option_id":"question-5-option-1"}
	]}`

	w := postJSON(t, newFixtureServer(), "/topics/intro/quiz/submissions", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	assertContentTypeJSON(t, w)

	result := decodeObject(t, w)
	assertExactKeys(t, result, []string{"score", "total", "answers"})
	if result["score"] != float64(1) || result["total"] != float64(4) {
		t.Errorf("expected score 1 of 4, got %v of %v", result["score"], result["total"])
	}

	answers, ok := result["answers"].([]any)
	if !ok || len(answers) != 4 {
		t.Fatalf("expected 4 graded answers, got %v", result["answers"])
	}
	want := []struct {
		questionID  string
		correct     bool
		correctText string
		explanation string
	}{
		{"question-1", true, "First option", "Because 1."},
		{"question-2", false, "First option", "Because 2."},
		{"question-3", false, "First option", "Because 3."},
		{"question-4", false, "First option", "Because 4."},
	}
	for i, raw := range answers {
		answer, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected an answer object, got %v", raw)
		}
		assertExactKeys(t, answer, []string{"question_id", "is_correct", "correct_option_text", "explanation"})
		if answer["question_id"] != want[i].questionID {
			t.Errorf("answer %d: expected question %q, got %v", i, want[i].questionID, answer["question_id"])
		}
		if answer["is_correct"] != want[i].correct {
			t.Errorf("answer %d: expected is_correct %v, got %v", i, want[i].correct, answer["is_correct"])
		}
		if answer["correct_option_text"] != want[i].correctText {
			t.Errorf("answer %d: expected correct option text %q, got %v", i, want[i].correctText, answer["correct_option_text"])
		}
		if answer["explanation"] != want[i].explanation {
			t.Errorf("answer %d: expected explanation %q, got %v", i, want[i].explanation, answer["explanation"])
		}
	}
}

func TestSubmitQuizRejectsQuestionNotBelongingToTopic(t *testing.T) {
	body := `{"answers":[
		{"question_id":"question-1","option_id":"question-1-option-1"},
		{"question_id":"question-advanced","option_id":"question-advanced-option-1"}
	]}`

	w := postJSON(t, newFixtureServer(), "/topics/intro/quiz/submissions", body)

	assertErrorCode(t, w, http.StatusBadRequest, "validation_error")
	if strings.Contains(w.Body.String(), "score") {
		t.Errorf("rejected submission was partly graded: %s", w.Body.String())
	}
}

func TestSubmitQuizRejectsInvalidAnswerLists(t *testing.T) {
	for name, body := range map[string]string{
		"empty list":    `{"answers":[]}`,
		"missing field": `{}`,
		"invalid JSON":  `{"answers":[`,
	} {
		t.Run(name, func(t *testing.T) {
			w := postJSON(t, newFixtureServer(), "/topics/intro/quiz/submissions", body)
			assertErrorCode(t, w, http.StatusBadRequest, "validation_error")
		})
	}
}

func TestSubmitQuizWorksWithoutSession(t *testing.T) {
	body := `{"answers":[{"question_id":"question-1","option_id":"question-1-option-1"}]}`

	for name, cookie := range map[string]string{
		"no cookie":      "",
		"invalid cookie": "session=invalid-token",
		"expired cookie": "session=expired-token",
	} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/topics/intro/quiz/submissions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if cookie != "" {
				req.Header.Set("Cookie", cookie)
			}
			newFixtureServer().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}
			result := decodeObject(t, w)
			if result["score"] != float64(1) || result["total"] != float64(1) {
				t.Errorf("expected the anonymous submission to be graded, got %v", result)
			}
		})
	}
}

func TestSubmitQuizNotFound(t *testing.T) {
	body := `{"answers":[{"question_id":"question-1","option_id":null}]}`

	for _, slug := range []string{"unknown", "draft-topic"} {
		t.Run(slug, func(t *testing.T) {
			w := postJSON(t, newFixtureServer(), "/topics/"+slug+"/quiz/submissions", body)
			assertErrorCode(t, w, http.StatusNotFound, "not_found")
		})
	}
}

func TestSubmitQuizHidesRepositoryErrors(t *testing.T) {
	cause := errors.New("query failed for user 42: password_hash=secret")
	h := newContentTestServer(failingRepository{err: cause})
	body := `{"answers":[{"question_id":"question-1","option_id":null}]}`

	w := postJSON(t, h, "/topics/intro/quiz/submissions", body)

	assertErrorCode(t, w, http.StatusInternalServerError, "internal_error")
	for _, leak := range []string{"query failed", "user 42", "password_hash", "secret"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("error body leaks %q: %s", leak, w.Body.String())
		}
	}
}
