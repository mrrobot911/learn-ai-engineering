package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/middleware"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

type attemptTestServer struct {
	*authServer
}

func newAttemptTestServer(t *testing.T) *attemptTestServer {
	t.Helper()

	topics := []repo.Topic{
		{ID: "topic-1", Slug: "intro", Title: "Introduction", Position: 1, Status: repo.StatusPublished},
	}
	questions := []repo.Question{
		{ID: "question-1", TopicID: "topic-1", Text: "Question 1?", Explanation: "Because 1.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "question-1-option-1", Text: "First option", Position: 1, IsCorrect: true},
			{ID: "question-1-option-2", Text: "Second option", Position: 2},
		}},
		{ID: "question-2", TopicID: "topic-1", Text: "Question 2?", Explanation: "Because 2.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "question-2-option-1", Text: "First option", Position: 1, IsCorrect: true},
			{ID: "question-2-option-2", Text: "Second option", Position: 2},
		}},
	}

	authRepo := repo.NewMemoryAuthRepository()
	attemptRepo := repo.NewMemoryAttemptRepository()
	authService := service.NewAuthService(authRepo)
	contentService := service.NewContentService(repo.NewMemoryContentRepository(topics, nil, questions), attemptRepo)
	authHandler := handler.NewAuthHandler(authService, false)
	contentHandler := handler.NewContentHandler(contentService)

	withSession := middleware.Session(authService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.Handle("POST /topics/{topicSlug}/quiz/submissions", withSession(middleware.OptionalCSRF(http.HandlerFunc(contentHandler.SubmitQuiz))))
	mux.Handle("GET /me/attempts", withSession(http.HandlerFunc(contentHandler.ListMyAttempts)))
	mux.Handle("GET /me/attempts/{attemptId}", withSession(http.HandlerFunc(contentHandler.GetMyAttempt)))

	return &attemptTestServer{
		authServer: &authServer{handler: mux, authRepo: authRepo},
	}
}

func decodeAttemptList(t *testing.T, body []byte) []map[string]any {
	t.Helper()
	var list []map[string]any
	if err := json.Unmarshal(body, &list); err != nil {
		t.Fatalf("decode attempt list: %v; body: %s", err, body)
	}
	if list == nil {
		t.Fatalf("expected a JSON array, got null")
	}
	return list
}

const attemptSubmissionBody = `{"answers":[
	{"question_id":"question-1","option_id":"question-1-option-1"},
	{"question_id":"question-2","option_id":null}
]}`

func TestSubmitQuizStoresAttemptAndListsIt(t *testing.T) {
	s := newAttemptTestServer(t)
	_, csrf, cookie := registerUser(t, s.authServer, "user@example.com", "password")

	w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
		body:   attemptSubmissionBody,
		cookie: cookie.Value,
		csrf:   csrf,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	list := authRequest(t, s.handler, http.MethodGet, "/me/attempts", requestOptions{cookie: cookie.Value})
	if list.Code != http.StatusOK {
		t.Fatalf("list attempts: expected status 200, got %d: %s", list.Code, list.Body.String())
	}
	items := decodeAttemptList(t, list.Body.Bytes())
	if len(items) != 1 {
		t.Fatalf("expected one stored attempt, got %d: %v", len(items), items)
	}
	assertExactKeys(t, items[0], "id", "topic_id", "topic_slug", "score", "total", "created_at")
	if items[0]["topic_id"] != "topic-1" || items[0]["topic_slug"] != "intro" {
		t.Errorf("unexpected topic in the attempt list: %v", items[0])
	}
	if items[0]["score"] != float64(1) || items[0]["total"] != float64(2) {
		t.Errorf("expected score 1 of 2, got %v of %v", items[0]["score"], items[0]["total"])
	}

	attemptID, ok := items[0]["id"].(string)
	if !ok || attemptID == "" {
		t.Fatalf("expected an attempt id, got %v", items[0]["id"])
	}

	detail := authRequest(t, s.handler, http.MethodGet, "/me/attempts/"+attemptID, requestOptions{cookie: cookie.Value})
	if detail.Code != http.StatusOK {
		t.Fatalf("get attempt: expected status 200, got %d: %s", detail.Code, detail.Body.String())
	}
	body := decodeBody(t, detail)
	assertExactKeys(t, body, "id", "topic_id", "topic_slug", "score", "total", "created_at", "answers")
	answers, ok := body["answers"].([]any)
	if !ok || len(answers) != 2 {
		t.Fatalf("expected two snapshot answers, got %v", body["answers"])
	}

	first, ok := answers[0].(map[string]any)
	if !ok {
		t.Fatalf("expected an answer object, got %v", answers[0])
	}
	assertExactKeys(t, first, "question_id", "question_text", "selected_option_text", "correct_option_text", "is_correct")
	if first["question_id"] != "question-1" || first["question_text"] != "Question 1?" {
		t.Errorf("unexpected first snapshot question: %v", first)
	}
	if first["selected_option_text"] != "First option" || first["correct_option_text"] != "First option" || first["is_correct"] != true {
		t.Errorf("unexpected first snapshot answer: %v", first)
	}

	second, ok := answers[1].(map[string]any)
	if !ok {
		t.Fatalf("expected an answer object, got %v", answers[1])
	}
	if second["selected_option_text"] != "" || second["correct_option_text"] != "First option" || second["is_correct"] != false {
		t.Errorf("expected the unanswered snapshot to keep empty selected text, got %v", second)
	}
}

func TestAttemptsBelongOnlyToTheirUser(t *testing.T) {
	s := newAttemptTestServer(t)
	_, csrf, firstCookie := registerUser(t, s.authServer, "first@example.com", "password")
	_, _, secondCookie := registerUser(t, s.authServer, "second@example.com", "password")

	w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
		body:   attemptSubmissionBody,
		cookie: firstCookie.Value,
		csrf:   csrf,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("submit quiz: expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	items := decodeAttemptList(t, authRequest(t, s.handler, http.MethodGet, "/me/attempts", requestOptions{cookie: firstCookie.Value}).Body.Bytes())
	if len(items) != 1 {
		t.Fatalf("expected one attempt for the first user, got %v", items)
	}
	attemptID, _ := items[0]["id"].(string)

	otherList := authRequest(t, s.handler, http.MethodGet, "/me/attempts", requestOptions{cookie: secondCookie.Value})
	if otherList.Code != http.StatusOK {
		t.Fatalf("second user's list: expected status 200, got %d: %s", otherList.Code, otherList.Body.String())
	}
	if got := decodeAttemptList(t, otherList.Body.Bytes()); len(got) != 0 {
		t.Errorf("expected the second user's list to be empty, got %v", got)
	}

	otherDetail := authRequest(t, s.handler, http.MethodGet, "/me/attempts/"+attemptID, requestOptions{cookie: secondCookie.Value})
	assertErrorResponse(t, otherDetail, http.StatusNotFound, "not_found")
}

func TestAttemptHistoryRequiresASession(t *testing.T) {
	s := newAttemptTestServer(t)

	for _, path := range []string{"/me/attempts", "/me/attempts/attempt-1"} {
		t.Run(path, func(t *testing.T) {
			w := authRequest(t, s.handler, http.MethodGet, path, requestOptions{})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		})
	}
}
