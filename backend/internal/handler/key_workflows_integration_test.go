//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func seedWorkflowFixtures(t *testing.T, pool *pgxpool.Pool) (topicSlug, lessonID, draftLessonID, questionID, optionCorrect, optionWrong, foreignQuestionID string) {
	t.Helper()
	ctx := context.Background()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed (%s): %v", sql, err)
		}
	}
	topicSlug = "intro"
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		"topic-intro", topicSlug, "Introduction", "Start here.", 1, "published")
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		"topic-draft", "draft-topic", "Draft", "Hidden.", 2, "draft")
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		"topic-other", "other", "Other", "Other topic.", 3, "published")

	lessonID = "lesson-pub-1"
	draftLessonID = "lesson-draft-1"
	exec(`INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		lessonID, "topic-intro", "First lesson", "# Hello\n\nBody.", 1, "published")
	exec(`INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		draftLessonID, "topic-intro", "Draft lesson", "Hidden body.", 2, "draft")

	questionID = "question-1"
	optionCorrect = "option-1-correct"
	optionWrong = "option-1-wrong"
	exec(`INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ($1,$2,$3,$4,$5)`,
		questionID, "topic-intro", "What is 1+1?", "Because math.", "published")
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		optionCorrect, questionID, "2", 1, true, "published")
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		optionWrong, questionID, "3", 2, false, "published")
	exec(`INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ($1,$2,$3,$4,$5)`,
		"question-draft", "topic-intro", "Draft question?", "Hidden.", "draft")
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		"option-draft", "question-draft", "Nope", 1, true, "draft")

	foreignQuestionID = "question-foreign"
	exec(`INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ($1,$2,$3,$4,$5)`,
		foreignQuestionID, "topic-other", "Foreign?", "Other topic.", "published")
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		"option-foreign", foreignQuestionID, "Yes", 1, true, "published")
	return
}

func doJSON(t *testing.T, h http.Handler, method, path, body, cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	return authRequest(t, h, method, path, requestOptions{body: body, cookie: cookie, csrf: csrf})
}

func decodeAny(t *testing.T, w *httptest.ResponseRecorder) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode json: %v; body: %s", err, w.Body.String())
	}
	return v
}

func TestKeyWorkflows_EndToEnd(t *testing.T) {
	pool := integrationPool(t)
	resetIntegrationTables(t, pool)
	applyIntegrationMigrations(t, pool)
	topicSlug, lessonID, draftLessonID, questionID, optionCorrect, _, foreignQuestionID := seedWorkflowFixtures(t, pool)
	h := buildIntegrationHandler(t, pool)

	// Browsing published content (AC1) and unpublished never appears (AC2).
	t.Run("browse published content hides drafts", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/topics", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET /topics: %d %s", w.Code, w.Body.String())
		}
		topics, ok := decodeAny(t, w).([]any)
		if !ok || len(topics) != 2 {
			t.Fatalf("expected 2 published topics, got %#v", decodeAny(t, w))
		}
		for _, item := range topics {
			m := item.(map[string]any)
			if m["slug"] == "draft-topic" {
				t.Fatal("draft topic must not appear in public list")
			}
		}

		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/topics/"+topicSlug, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET topic: %d %s", w.Code, w.Body.String())
		}
		topic := decodeAny(t, w).(map[string]any)
		if topic["slug"] != topicSlug {
			t.Fatalf("unexpected topic: %#v", topic)
		}
		lessons, _ := topic["lessons"].([]any)
		if len(lessons) != 1 {
			t.Fatalf("expected 1 published lesson, got %#v", topic["lessons"])
		}
		if lessons[0].(map[string]any)["id"] != lessonID {
			t.Fatalf("unexpected lesson: %#v", lessons[0])
		}

		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/topics/draft-topic", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("draft topic should be 404, got %d %s", w.Code, w.Body.String())
		}

		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/"+lessonID, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET lesson: %d %s", w.Code, w.Body.String())
		}
		lesson := decodeAny(t, w).(map[string]any)
		if lesson["id"] != lessonID || lesson["body"] != "# Hello\n\nBody." {
			t.Fatalf("unexpected lesson body preserved: %#v", lesson)
		}

		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/"+draftLessonID, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("draft lesson should be 404, got %d %s", w.Code, w.Body.String())
		}
	})

	// Register, login, session, logout (AC1).
	var sessionID, csrfToken string
	t.Run("register login session logout", func(t *testing.T) {
		w := doJSON(t, h, http.MethodPost, "/auth/register", `{"email":"user@example.com","password":"password123"}`, "", "")
		if w.Code != http.StatusCreated {
			t.Fatalf("register: %d %s", w.Code, w.Body.String())
		}
		sessionID = sessionCookieValue(t, w)
		body := decodeBody(t, w)
		csrfToken, _ = body["csrf_token"].(string)
		if csrfToken == "" || sessionID == "" {
			t.Fatalf("register missing session/csrf: %#v", body)
		}

		w = doJSON(t, h, http.MethodGet, "/auth/session", "", sessionID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("session: %d %s", w.Code, w.Body.String())
		}

		w = doJSON(t, h, http.MethodPost, "/auth/logout", "", sessionID, csrfToken)
		if w.Code != http.StatusOK {
			t.Fatalf("logout: %d %s", w.Code, w.Body.String())
		}

		w = doJSON(t, h, http.MethodPost, "/auth/login", `{"email":"user@example.com","password":"password123"}`, "", "")
		if w.Code != http.StatusOK {
			t.Fatalf("login: %d %s", w.Code, w.Body.String())
		}
		sessionID = sessionCookieValue(t, w)
		csrfToken, _ = decodeBody(t, w)["csrf_token"].(string)
		if csrfToken == "" {
			t.Fatal("login missing csrf_token")
		}
	})

	// Start and submit quiz (AC1), foreign question rejected (AC2), draft not in quiz.
	t.Run("quiz start submit and foreign question rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/topics/"+topicSlug+"/quiz", strings.NewReader("")))
		if w.Code != http.StatusOK {
			t.Fatalf("start quiz: %d %s", w.Code, w.Body.String())
		}
		qs, ok := decodeAny(t, w).([]any)
		if !ok || len(qs) != 1 {
			t.Fatalf("expected 1 published question, got %#v", decodeAny(t, w))
		}
		q := qs[0].(map[string]any)
		if q["id"] != questionID {
			t.Fatalf("unexpected question: %#v", q)
		}
		if _, has := q["correct_option_id"]; has {
			t.Fatalf("quiz must not expose correct option before submission: %#v", q)
		}
		if _, has := q["is_correct"]; has {
			t.Fatalf("quiz must not expose correctness: %#v", q)
		}
		raw := w.Body.String()
		if strings.Contains(raw, "is_correct") || strings.Contains(raw, optionCorrect) {
			// optionCorrect is an ID - IDs may appear as option ids; is_correct must not.
			if strings.Contains(raw, "is_correct") {
				t.Fatalf("quiz response leaked correctness: %s", raw)
			}
		}

		// Foreign question rejected (AC2) - authenticated submission.
		foreignBody := `{"answers":[{"question_id":"` + foreignQuestionID + `","option_id":"option-foreign"}]}`
		w = doJSON(t, h, http.MethodPost, "/topics/"+topicSlug+"/quiz/submissions", foreignBody, sessionID, csrfToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("foreign question should be 400, got %d %s", w.Code, w.Body.String())
		}
		errBody := decodeBody(t, w)
		if errBody["code"] != "validation_error" {
			t.Fatalf("expected validation_error, got %#v", errBody)
		}

		// Valid submission stores attempt (authenticated).
		subBody := `{"answers":[{"question_id":"` + questionID + `","option_id":"` + optionCorrect + `"}]}`
		w = doJSON(t, h, http.MethodPost, "/topics/"+topicSlug+"/quiz/submissions", subBody, sessionID, csrfToken)
		if w.Code != http.StatusOK {
			t.Fatalf("submit quiz: %d %s", w.Code, w.Body.String())
		}
		result := decodeBody(t, w)
		if result["score"] != float64(1) || result["total"] != float64(1) {
			t.Fatalf("unexpected result: %#v", result)
		}
	})

	// Saving progress (AC1).
	t.Run("save and read progress", func(t *testing.T) {
		w := doJSON(t, h, http.MethodPut, "/me/progress/lessons/"+lessonID, "", sessionID, csrfToken)
		if w.Code != http.StatusOK {
			t.Fatalf("mark complete: %d %s", w.Code, w.Body.String())
		}
		w = doJSON(t, h, http.MethodGet, "/me/progress", "", sessionID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("get progress: %d %s", w.Code, w.Body.String())
		}
		ids, ok := decodeAny(t, w).([]any)
		if !ok || len(ids) != 1 || ids[0] != lessonID {
			t.Fatalf("expected progress [%s], got %#v", lessonID, decodeAny(t, w))
		}

		w = doJSON(t, h, http.MethodDelete, "/me/progress/lessons/"+lessonID, "", sessionID, csrfToken)
		if w.Code != http.StatusOK {
			t.Fatalf("unmark complete: %d %s", w.Code, w.Body.String())
		}
		w = doJSON(t, h, http.MethodGet, "/me/progress", "", sessionID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("get progress after delete: %d %s", w.Code, w.Body.String())
		}
		ids, _ = decodeAny(t, w).([]any)
		if len(ids) != 0 {
			t.Fatalf("expected empty progress, got %#v", ids)
		}
	})

	// Reading quiz history (AC1).
	t.Run("read quiz history", func(t *testing.T) {
		w := doJSON(t, h, http.MethodGet, "/me/attempts", "", sessionID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("list attempts: %d %s", w.Code, w.Body.String())
		}
		attempts, ok := decodeAny(t, w).([]any)
		if !ok || len(attempts) < 1 {
			t.Fatalf("expected at least 1 attempt, got %#v", decodeAny(t, w))
		}
		first := attempts[0].(map[string]any)
		attemptID, _ := first["id"].(string)
		if attemptID == "" {
			t.Fatalf("attempt missing id: %#v", first)
		}
		if first["topic_slug"] != topicSlug {
			t.Fatalf("unexpected attempt topic: %#v", first)
		}

		w = doJSON(t, h, http.MethodGet, "/me/attempts/"+attemptID, "", sessionID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("get attempt: %d %s", w.Code, w.Body.String())
		}
		detail := decodeBody(t, w)
		if detail["id"] != attemptID {
			t.Fatalf("unexpected attempt detail: %#v", detail)
		}
		answers, _ := detail["answers"].([]any)
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer in detail, got %#v", detail["answers"])
		}
	})
}

// TestKeyWorkflows_CanRunTwice proves AC3 by running the full setup+assertions
// pattern in a second independent test that shares the same database.
func TestKeyWorkflows_CanRunTwice(t *testing.T) {
	pool := integrationPool(t)
	resetIntegrationTables(t, pool)
	applyIntegrationMigrations(t, pool)
	topicSlug, _, _, questionID, optionCorrect, _, _ := seedWorkflowFixtures(t, pool)
	h := buildIntegrationHandler(t, pool)

	w := doJSON(t, h, http.MethodPost, "/auth/register", `{"email":"twice@example.com","password":"password123"}`, "", "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	sessionID := sessionCookieValue(t, w)
	csrf, _ := decodeBody(t, w)["csrf_token"].(string)

	subBody := `{"answers":[{"question_id":"` + questionID + `","option_id":"` + optionCorrect + `"}]}`
	w = doJSON(t, h, http.MethodPost, "/topics/"+topicSlug+"/quiz/submissions", subBody, sessionID, csrf)
	if w.Code != http.StatusOK {
		t.Fatalf("submit: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodGet, "/me/attempts", "", sessionID, "")
	if w.Code != http.StatusOK {
		t.Fatalf("attempts: %d %s", w.Code, w.Body.String())
	}
	attempts, _ := decodeAny(t, w).([]any)
	if len(attempts) != 1 {
		t.Fatalf("expected 1 attempt on second run, got %d", len(attempts))
	}
}
