package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
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

const quizBody = `{"answers":[{"question_id":"question-1","option_id":"question-1-option-1"}]}`

type authServer struct {
	handler  http.Handler
	authRepo *repo.MemoryAuthRepository
}

func buildAuthHandler(authRepo repo.AuthRepository, secureCookie bool) http.Handler {
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService, secureCookie)

	topics := []repo.Topic{
		{ID: "topic-1", Slug: "intro", Title: "Introduction", Position: 1, Status: repo.StatusPublished},
	}
	questions := []repo.Question{
		{
			ID:          "question-1",
			TopicID:     "topic-1",
			Text:        "Question 1?",
			Explanation: "Because 1.",
			Status:      repo.StatusPublished,
			Options: []repo.Option{
				{ID: "question-1-option-1", Text: "First option", Position: 1, IsCorrect: true},
				{ID: "question-1-option-2", Text: "Second option", Position: 2},
			},
		},
	}
	contentHandler := handler.NewContentHandler(service.NewContentService(repo.NewMemoryContentRepository(topics, nil, questions)))

	withSession := middleware.Session(authService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /auth/session", withSession(http.HandlerFunc(authHandler.CurrentSession)))
	mux.Handle("POST /auth/logout", withSession(middleware.CSRF(http.HandlerFunc(authHandler.Logout))))
	mux.Handle("POST /topics/{topicSlug}/quiz/submissions", withSession(middleware.OptionalCSRF(http.HandlerFunc(contentHandler.SubmitQuiz))))
	return mux
}

func newAuthTestServer(t *testing.T, secureCookie bool) *authServer {
	t.Helper()
	authRepo := repo.NewMemoryAuthRepository()
	return &authServer{handler: buildAuthHandler(authRepo, secureCookie), authRepo: authRepo}
}

type requestOptions struct {
	body   string
	cookie string
	csrf   string
}

func authRequest(t *testing.T, h http.Handler, method, path string, opts requestOptions) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if opts.body == "" {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(opts.body)
	}
	req := httptest.NewRequest(method, path, body)
	if opts.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if opts.cookie != "" {
		req.Header.Set("Cookie", "session="+opts.cookie)
	}
	if opts.csrf != "" {
		req.Header.Set("X-CSRF-Token", opts.csrf)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v; body: %s", err, w.Body.String())
	}
	return body
}

func assertJSON(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", got)
	}
}

func assertExactKeys(t *testing.T, object map[string]any, want ...string) {
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

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Errorf("expected status %d, got %d: %s", status, w.Code, w.Body.String())
	}
	assertJSON(t, w)
	body := decodeBody(t, w)
	assertExactKeys(t, body, "code", "message")
	if body["code"] != code {
		t.Errorf("expected code %q, got %v", code, body["code"])
	}
	if message, ok := body["message"].(string); !ok || message == "" {
		t.Errorf("expected a non-empty message, got %v", body["message"])
	}
}

func assertUserAndToken(t *testing.T, w *httptest.ResponseRecorder, email string) (string, string) {
	t.Helper()
	assertJSON(t, w)
	body := decodeBody(t, w)
	assertExactKeys(t, body, "user", "csrf_token")

	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected a user object, got %v", body["user"])
	}
	assertExactKeys(t, user, "id", "email")
	id, ok := user["id"].(string)
	if !ok || id == "" {
		t.Errorf("expected a non-empty user id, got %v", user["id"])
	}
	if user["email"] != email {
		t.Errorf("expected email %q, got %v", email, user["email"])
	}

	token, ok := body["csrf_token"].(string)
	if !ok || token == "" {
		t.Errorf("expected a non-empty csrf_token, got %v", body["csrf_token"])
	}
	return id, token
}

func responseCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}
	t.Fatalf("expected a session Set-Cookie header, got %v", w.Header().Values("Set-Cookie"))
	return nil
}

func hasSessionCookie(w *httptest.ResponseRecorder) bool {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session" {
			return true
		}
	}
	return false
}

func assertSessionCookie(t *testing.T, cookie *http.Cookie, secure bool) {
	t.Helper()
	if cookie.Value == "" {
		t.Error("expected a non-empty session cookie value")
	}
	if !cookie.HttpOnly {
		t.Error("expected an HttpOnly session cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
	if cookie.Path != "/" {
		t.Errorf("expected path /, got %q", cookie.Path)
	}
	if cookie.Secure != secure {
		t.Errorf("expected Secure=%v, got %v", secure, cookie.Secure)
	}
	if cookie.MaxAge != int(service.SessionDuration.Seconds()) {
		t.Errorf("expected Max-Age %d, got %d", int(service.SessionDuration.Seconds()), cookie.MaxAge)
	}
	diff := cookie.Expires.Sub(time.Now().Add(service.SessionDuration))
	if diff < -2*time.Minute || diff > 2*time.Minute {
		t.Errorf("expected an expiry about 7 days away, got %s", cookie.Expires)
	}
}

func assertClearedCookie(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	cookie := responseCookie(t, w)
	if cookie.Value != "" {
		t.Errorf("expected an empty cookie value, got %q", cookie.Value)
	}
	if cookie.MaxAge >= 0 && !cookie.Expires.Before(time.Now()) {
		t.Errorf("expected Max-Age=0 or an expiry in the past, got Max-Age=%d expires=%s", cookie.MaxAge, cookie.Expires)
	}
}

func registerUser(t *testing.T, s *authServer, email, password string) (string, string, *http.Cookie) {
	t.Helper()
	w := authRequest(t, s.handler, http.MethodPost, "/auth/register", requestOptions{
		body: `{"email":"` + email + `","password":"` + password + `"}`,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register %s: expected status 201, got %d: %s", email, w.Code, w.Body.String())
	}
	_, token := assertUserAndToken(t, w, email)
	cookie := responseCookie(t, w)
	return cookie.Value, token, cookie
}

func TestRegisterReturnsSessionAndAcceptsTheToken(t *testing.T) {
	for _, length := range []int{8, 128} {
		t.Run(strings.Repeat("x", length), func(t *testing.T) {
			s := newAuthTestServer(t, false)
			password := strings.Repeat("a", length)
			email := "user@example.com"

			w := authRequest(t, s.handler, http.MethodPost, "/auth/register", requestOptions{
				body: `{"email":"` + email + `","password":"` + password + `"}`,
			})
			if w.Code != http.StatusCreated {
				t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
			}
			userID, token := assertUserAndToken(t, w, email)
			cookie := responseCookie(t, w)
			assertSessionCookie(t, cookie, false)

			if strings.Contains(w.Body.String(), password) {
				t.Error("the registration response contains the plaintext password")
			}

			storedUser, err := s.authRepo.GetUserByEmail(context.Background(), email)
			if err != nil {
				t.Fatalf("get stored user: %v", err)
			}
			if storedUser.ID != userID {
				t.Errorf("expected the response user id %q, got %q", userID, storedUser.ID)
			}
			if !strings.HasPrefix(storedUser.PasswordHash, "$argon2id$") {
				t.Errorf("expected an Argon2id hash, got %q", storedUser.PasswordHash)
			}
			if strings.Contains(storedUser.PasswordHash, password) {
				t.Error("the stored hash contains the plaintext password")
			}

			storedSession, err := s.authRepo.GetSessionByID(context.Background(), cookie.Value)
			if err != nil {
				t.Fatalf("get stored session: %v", err)
			}
			if got := storedSession.ExpiresAt.Sub(storedSession.CreatedAt); got != service.SessionDuration {
				t.Errorf("expected a %s session, got %s", service.SessionDuration, got)
			}

			read := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: cookie.Value})
			if read.Code != http.StatusOK {
				t.Fatalf("session read: expected status 200, got %d: %s", read.Code, read.Body.String())
			}
			if gotID, gotToken := assertUserAndToken(t, read, email); gotID != userID || gotToken != token {
				t.Errorf("session read returned a different user or token")
			}

			logout := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
				cookie: cookie.Value,
				csrf:   token,
			})
			if logout.Code != http.StatusOK {
				t.Fatalf("logout with the issued token: expected status 200, got %d: %s", logout.Code, logout.Body.String())
			}
		})
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"missing email", "", "password"},
		{"invalid email", "not-an-email", "password"},
		{"missing password", "user@example.com", ""},
		{"password too short", "user@example.com", strings.Repeat("a", 7)},
		{"password too long", "user@example.com", strings.Repeat("a", 129)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newAuthTestServer(t, false)
			w := authRequest(t, s.handler, http.MethodPost, "/auth/register", requestOptions{
				body: `{"email":"` + tt.email + `","password":"` + tt.password + `"}`,
			})

			assertErrorResponse(t, w, http.StatusBadRequest, "validation_error")
			if hasSessionCookie(w) {
				t.Error("a rejected registration set a session cookie")
			}
			if got := s.authRepo.UserCount(); got != 0 {
				t.Errorf("expected no user, got %d", got)
			}
			if got := s.authRepo.SessionCount(); got != 0 {
				t.Errorf("expected no session, got %d", got)
			}
		})
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	s := newAuthTestServer(t, false)
	registerUser(t, s, "user@example.com", "password")

	before, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("get stored user: %v", err)
	}
	users, sessions := s.authRepo.UserCount(), s.authRepo.SessionCount()

	w := authRequest(t, s.handler, http.MethodPost, "/auth/register", requestOptions{
		body: `{"email":"user@example.com","password":"another-password"}`,
	})

	assertErrorResponse(t, w, http.StatusConflict, "conflict")
	if hasSessionCookie(w) {
		t.Error("a duplicate registration set a session cookie")
	}
	if got := s.authRepo.UserCount(); got != users {
		t.Errorf("expected %d users, got %d", users, got)
	}
	if got := s.authRepo.SessionCount(); got != sessions {
		t.Errorf("expected %d sessions, got %d", sessions, got)
	}
	after, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("get stored user after duplicate: %v", err)
	}
	if after.PasswordHash != before.PasswordHash {
		t.Error("the duplicate registration changed the stored password hash")
	}
}

func TestRegisterStoresDifferentSaltedHashes(t *testing.T) {
	s := newAuthTestServer(t, false)
	registerUser(t, s, "first@example.com", "password")
	registerUser(t, s, "second@example.com", "password")

	first, err := s.authRepo.GetUserByEmail(context.Background(), "first@example.com")
	if err != nil {
		t.Fatalf("get first user: %v", err)
	}
	second, err := s.authRepo.GetUserByEmail(context.Background(), "second@example.com")
	if err != nil {
		t.Fatalf("get second user: %v", err)
	}
	if first.PasswordHash == second.PasswordHash {
		t.Error("two users with the same password share a hash, so the salt is missing")
	}
}

func TestLoginReturnsSessionCookieAndAcceptsToken(t *testing.T) {
	s := newAuthTestServer(t, false)
	registerUser(t, s, "user@example.com", "password")

	w := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
		body: `{"email":"user@example.com","password":"password"}`,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	_, token := assertUserAndToken(t, w, "user@example.com")
	cookie := responseCookie(t, w)
	assertSessionCookie(t, cookie, false)

	stored, err := s.authRepo.GetSessionByID(context.Background(), cookie.Value)
	if err != nil {
		t.Fatalf("get stored session: %v", err)
	}
	if got := stored.ExpiresAt.Sub(stored.CreatedAt); got != service.SessionDuration {
		t.Errorf("expected a %s session, got %s", service.SessionDuration, got)
	}

	logout := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
		cookie: cookie.Value,
		csrf:   token,
	})
	if logout.Code != http.StatusOK {
		t.Fatalf("logout with the login token: expected status 200, got %d: %s", logout.Code, logout.Body.String())
	}
}

func TestLoginCookieIsSecureInProductionOnly(t *testing.T) {
	tests := []struct {
		name   string
		secure bool
	}{
		{"development", false},
		{"production", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newAuthTestServer(t, tt.secure)
			registerUser(t, s, "user@example.com", "password")

			w := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
				body: `{"email":"user@example.com","password":"password"}`,
			})
			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}
			assertSessionCookie(t, responseCookie(t, w), tt.secure)
		})
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", "user@example.com", "wrong"},
		{"wrong short password", "user@example.com", "x"},
		{"wrong long password", "user@example.com", strings.Repeat("x", 200)},
		{"unknown email", "unknown@example.com", "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newAuthTestServer(t, false)
			registerUser(t, s, "user@example.com", "password")
			sessions := s.authRepo.SessionCount()

			w := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
				body: `{"email":"` + tt.email + `","password":"` + tt.password + `"}`,
			})

			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
			if hasSessionCookie(w) {
				t.Error("a failed login set a session cookie")
			}
			if strings.Contains(w.Body.String(), "csrf_token") {
				t.Error("a failed login returned a CSRF token")
			}
			if got := s.authRepo.SessionCount(); got != sessions {
				t.Errorf("expected %d sessions after the failed login, got %d", sessions, got)
			}
		})
	}
}

func TestLoginRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"missing email", "", "password"},
		{"invalid email", "not-an-email", "password"},
		{"missing password", "user@example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newAuthTestServer(t, false)
			registerUser(t, s, "user@example.com", "password")
			sessions := s.authRepo.SessionCount()

			w := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
				body: `{"email":"` + tt.email + `","password":"` + tt.password + `"}`,
			})

			assertErrorResponse(t, w, http.StatusBadRequest, "validation_error")
			if hasSessionCookie(w) {
				t.Error("an invalid login set a session cookie")
			}
			if got := s.authRepo.SessionCount(); got != sessions {
				t.Errorf("expected %d sessions after the invalid login, got %d", sessions, got)
			}
		})
	}
}

func TestSessionRead(t *testing.T) {
	s := newAuthTestServer(t, false)
	sessionID, token, _ := registerUser(t, s, "user@example.com", "password")

	t.Run("valid cookie without csrf header", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: sessionID})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if _, gotToken := assertUserAndToken(t, w, "user@example.com"); gotToken != token {
			t.Errorf("expected the session token, got %q", gotToken)
		}
	})

	t.Run("valid cookie with a wrong csrf header", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{
			cookie: sessionID,
			csrf:   "wrong-token",
		})
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing cookie", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{})
		assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
	})

	t.Run("unknown session", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: "unknown-session"})
		assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
	})

	t.Run("expired session", func(t *testing.T) {
		user, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		expired := repo.Session{
			ID:        "expired-session",
			UserID:    user.ID,
			CSRFToken: "expired-token",
			CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		if err := s.authRepo.CreateSession(context.Background(), expired); err != nil {
			t.Fatalf("create expired session: %v", err)
		}

		w := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: expired.ID})
		assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
	})
}

func TestLogoutRequiresTheCSRFToken(t *testing.T) {
	for _, tt := range []struct {
		name  string
		token string
	}{
		{"missing token", ""},
		{"wrong token", "wrong-token"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newAuthTestServer(t, false)
			sessionID, _, _ := registerUser(t, s, "user@example.com", "password")
			sessions := s.authRepo.SessionCount()

			w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
				cookie: sessionID,
				csrf:   tt.token,
			})

			assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
			if hasSessionCookie(w) {
				t.Error("a rejected logout cleared the cookie")
			}
			if got := s.authRepo.SessionCount(); got != sessions {
				t.Errorf("expected %d sessions after the rejected logout, got %d", sessions, got)
			}

			read := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: sessionID})
			if read.Code != http.StatusOK {
				t.Fatalf("the session should stay valid, got %d: %s", read.Code, read.Body.String())
			}
			_, freshToken := assertUserAndToken(t, read, "user@example.com")
			if freshToken == "" {
				t.Fatal("expected a fresh CSRF token")
			}

			logout := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
				cookie: sessionID,
				csrf:   freshToken,
			})
			if logout.Code != http.StatusOK {
				t.Fatalf("logout with the fresh token: expected status 200, got %d: %s", logout.Code, logout.Body.String())
			}
			assertJSON(t, logout)
			body := decodeBody(t, logout)
			assertExactKeys(t, body, "ok")
			if body["ok"] != true {
				t.Errorf("expected {\"ok\":true}, got %v", body)
			}
			assertClearedCookie(t, logout)

			if got := s.authRepo.SessionCount(); got != 0 {
				t.Errorf("expected no session after logout, got %d", got)
			}
			old := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: sessionID})
			assertErrorResponse(t, old, http.StatusUnauthorized, "authentication_error")
		})
	}
}

func TestCSRFTokenIsBoundToItsSession(t *testing.T) {
	s := newAuthTestServer(t, false)
	firstCookie, firstToken, _ := registerUser(t, s, "first@example.com", "password")

	login := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
		body: `{"email":"first@example.com","password":"password"}`,
	})
	if login.Code != http.StatusOK {
		t.Fatalf("login: expected status 200, got %d: %s", login.Code, login.Body.String())
	}
	_, secondToken := assertUserAndToken(t, login, "first@example.com")
	secondCookie := responseCookie(t, login)

	if firstToken == secondToken {
		t.Error("two sessions share a CSRF token")
	}

	t.Run("token of another session", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
			cookie: secondCookie.Value,
			csrf:   firstToken,
		})
		assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
	})

	t.Run("token in the body", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
			cookie: firstCookie,
			body:   `{"csrf_token":"` + firstToken + `"}`,
		})
		assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
	})

	t.Run("token in the query", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout?csrf_token="+firstToken, requestOptions{
			cookie: firstCookie,
		})
		assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
	})

	t.Run("token in the header for its own session", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
			cookie: firstCookie,
			csrf:   firstToken,
		})
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestLogoutRejectsInvalidSessions(t *testing.T) {
	s := newAuthTestServer(t, false)
	registerUser(t, s, "user@example.com", "password")

	t.Run("missing cookie with a token", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{csrf: "some-token"})
		assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
	})

	t.Run("missing cookie without a token", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{})
		assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
	})

	t.Run("unknown session", func(t *testing.T) {
		for _, token := range []string{"", "some-token"} {
			w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
				cookie: "unknown-session",
				csrf:   token,
			})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		}
	})

	t.Run("expired session", func(t *testing.T) {
		user, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		expired := repo.Session{
			ID:        "expired-session",
			UserID:    user.ID,
			CSRFToken: "expired-token",
			CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		if err := s.authRepo.CreateSession(context.Background(), expired); err != nil {
			t.Fatalf("create expired session: %v", err)
		}
		before := s.authRepo.SessionCount()

		for _, token := range []string{"", "expired-token"} {
			w := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
				cookie: expired.ID,
				csrf:   token,
			})
			assertErrorResponse(t, w, http.StatusUnauthorized, "authentication_error")
		}
		if got := s.authRepo.SessionCount(); got != before {
			t.Errorf("a rejected logout created or removed sessions: before %d, after %d", before, got)
		}
	})
}

func TestSessionIDIsOnlyReadFromTheCookie(t *testing.T) {
	s := newAuthTestServer(t, false)
	sessionID, token, _ := registerUser(t, s, "user@example.com", "password")

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		cookie string
		csrf   string
		status int
	}{
		{"session read with the id in the query", http.MethodGet, "/auth/session?session=" + sessionID, "", "", "", http.StatusUnauthorized},
		{"session read with the id in the body", http.MethodGet, "/auth/session", `{"session":"` + sessionID + `"}`, "", "", http.StatusUnauthorized},
		{"logout with the id in the query", http.MethodPost, "/auth/logout?session=" + sessionID, "", "", token, http.StatusUnauthorized},
		{"logout with the id in the body", http.MethodPost, "/auth/logout", `{"session":"` + sessionID + `"}`, "", token, http.StatusUnauthorized},
		{"session read with only the cookie", http.MethodGet, "/auth/session", "", sessionID, "", http.StatusOK},
		{"logout with only the cookie", http.MethodPost, "/auth/logout", "", sessionID, token, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := authRequest(t, s.handler, tt.method, tt.path, requestOptions{
				body:   tt.body,
				cookie: tt.cookie,
				csrf:   tt.csrf,
			})
			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
}

func TestQuizSubmissionRequiresCSRFOnlyForASession(t *testing.T) {
	s := newAuthTestServer(t, false)
	sessionID, token, _ := registerUser(t, s, "user@example.com", "password")

	anonymous := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{body: quizBody})
	if anonymous.Code != http.StatusOK {
		t.Fatalf("anonymous submission: expected status 200, got %d: %s", anonymous.Code, anonymous.Body.String())
	}

	for _, tt := range []struct {
		name   string
		cookie string
	}{
		{"no cookie", ""},
		{"invalid cookie", "invalid-session"},
		{"expired cookie", "expired-session"},
	} {
		t.Run("anonymous "+tt.name, func(t *testing.T) {
			if tt.cookie == "expired-session" {
				user, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
				if err != nil {
					t.Fatalf("get user: %v", err)
				}
				expired := repo.Session{
					ID:        "expired-session",
					UserID:    user.ID,
					CSRFToken: "expired-token",
					CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
					ExpiresAt: time.Now().Add(-time.Hour),
				}
				if err := s.authRepo.CreateSession(context.Background(), expired); err != nil {
					t.Fatalf("create expired session: %v", err)
				}
			}

			w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
				body:   quizBody,
				cookie: tt.cookie,
			})
			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}
			if w.Body.String() != anonymous.Body.String() {
				t.Errorf("expected the same graded result as an anonymous submission, got %s", w.Body.String())
			}
		})
	}

	for _, tt := range []struct {
		name string
		csrf string
	}{
		{"missing token", ""},
		{"wrong token", "wrong-token"},
	} {
		t.Run("with a session and "+tt.name, func(t *testing.T) {
			w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
				body:   quizBody,
				cookie: sessionID,
				csrf:   tt.csrf,
			})
			assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
		})
	}

	t.Run("with a session and its token", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
			body:   quizBody,
			cookie: sessionID,
			csrf:   token,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if w.Body.String() != anonymous.Body.String() {
			t.Errorf("expected the same graded result as an anonymous submission, got %s", w.Body.String())
		}
	})

	t.Run("with a session and the token only in the query", func(t *testing.T) {
		w := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions?csrf_token="+token, requestOptions{
			body:   quizBody,
			cookie: sessionID,
		})
		assertErrorResponse(t, w, http.StatusForbidden, "forbidden")
	})
}

type failingAuthRepository struct {
	*repo.MemoryAuthRepository
	err error

	failCreateUser     bool
	failGetUserByEmail bool
	failGetSessionByID bool
	failDeleteSession  bool
}

func (r *failingAuthRepository) CreateUser(ctx context.Context, user repo.User) error {
	if r.failCreateUser {
		return r.err
	}
	return r.MemoryAuthRepository.CreateUser(ctx, user)
}

func (r *failingAuthRepository) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	if r.failGetUserByEmail {
		return repo.User{}, r.err
	}
	return r.MemoryAuthRepository.GetUserByEmail(ctx, email)
}

func (r *failingAuthRepository) GetSessionByID(ctx context.Context, id string) (repo.Session, error) {
	if r.failGetSessionByID {
		return repo.Session{}, r.err
	}
	return r.MemoryAuthRepository.GetSessionByID(ctx, id)
}

func (r *failingAuthRepository) DeleteSession(ctx context.Context, id string) error {
	if r.failDeleteSession {
		return r.err
	}
	return r.MemoryAuthRepository.DeleteSession(ctx, id)
}

func assertSafeInternalError(t *testing.T, w *httptest.ResponseRecorder, secrets ...string) {
	t.Helper()
	assertErrorResponse(t, w, http.StatusInternalServerError, "internal_error")
	for _, secret := range secrets {
		if strings.Contains(w.Body.String(), secret) {
			t.Errorf("the error body leaks %q: %s", secret, w.Body.String())
		}
	}
}

func TestAuthEndpointsHideRepositoryErrors(t *testing.T) {
	cause := errors.New("query failed for user 42: password_hash=secret token=secret-token")

	t.Run("register", func(t *testing.T) {
		repository := &failingAuthRepository{MemoryAuthRepository: repo.NewMemoryAuthRepository(), err: cause, failCreateUser: true}
		h := buildAuthHandler(repository, false)

		w := authRequest(t, h, http.MethodPost, "/auth/register", requestOptions{
			body: `{"email":"user@example.com","password":"password"}`,
		})
		assertSafeInternalError(t, w, "query failed", "user 42", "password_hash", "secret", "token", "password")
	})

	t.Run("login", func(t *testing.T) {
		repository := &failingAuthRepository{MemoryAuthRepository: repo.NewMemoryAuthRepository(), err: cause, failGetUserByEmail: true}
		h := buildAuthHandler(repository, false)

		w := authRequest(t, h, http.MethodPost, "/auth/login", requestOptions{
			body: `{"email":"user@example.com","password":"password"}`,
		})
		assertSafeInternalError(t, w, "query failed", "user 42", "password_hash", "secret", "token", "password")
	})

	t.Run("session", func(t *testing.T) {
		repository := &failingAuthRepository{MemoryAuthRepository: repo.NewMemoryAuthRepository(), err: cause, failGetSessionByID: true}
		h := buildAuthHandler(repository, false)

		w := authRequest(t, h, http.MethodGet, "/auth/session", requestOptions{cookie: "some-session"})
		assertSafeInternalError(t, w, "query failed", "user 42", "password_hash", "secret", "token", "session")
	})

	t.Run("logout", func(t *testing.T) {
		memory := repo.NewMemoryAuthRepository()
		repository := &failingAuthRepository{MemoryAuthRepository: memory, err: cause, failDeleteSession: true}
		h := buildAuthHandler(repository, false)

		register := authRequest(t, h, http.MethodPost, "/auth/register", requestOptions{
			body: `{"email":"user@example.com","password":"password"}`,
		})
		if register.Code != http.StatusCreated {
			t.Fatalf("register: expected status 201, got %d: %s", register.Code, register.Body.String())
		}
		_, token := assertUserAndToken(t, register, "user@example.com")
		cookie := responseCookie(t, register)

		w := authRequest(t, h, http.MethodPost, "/auth/logout", requestOptions{cookie: cookie.Value, csrf: token})
		assertSafeInternalError(t, w, "query failed", "user 42", "password_hash", "secret", "token", "session")
	})

	t.Run("quiz submission", func(t *testing.T) {
		repository := &failingAuthRepository{MemoryAuthRepository: repo.NewMemoryAuthRepository(), err: cause, failGetSessionByID: true}
		h := buildAuthHandler(repository, false)

		w := authRequest(t, h, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
			body:   quizBody,
			cookie: "some-session",
		})
		assertSafeInternalError(t, w, "query failed", "user 42", "password_hash", "secret", "token", "session")
	})
}

func TestAuthFlowsDoNotLogSecrets(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)

	s := newAuthTestServer(t, false)
	s.handler = middleware.Logging(s.handler)
	password := "correct-horse-battery"

	sessionID, token, _ := registerUser(t, s, "user@example.com", password)
	user, err := s.authRepo.GetUserByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("get stored user: %v", err)
	}

	login := authRequest(t, s.handler, http.MethodPost, "/auth/login", requestOptions{
		body: `{"email":"user@example.com","password":"` + password + `"}`,
	})
	if login.Code != http.StatusOK {
		t.Fatalf("login: expected status 200, got %d: %s", login.Code, login.Body.String())
	}
	_, loginToken := assertUserAndToken(t, login, "user@example.com")
	loginCookie := responseCookie(t, login)

	read := authRequest(t, s.handler, http.MethodGet, "/auth/session", requestOptions{cookie: sessionID})
	if read.Code != http.StatusOK {
		t.Fatalf("session read: expected status 200, got %d: %s", read.Code, read.Body.String())
	}

	quiz := authRequest(t, s.handler, http.MethodPost, "/topics/intro/quiz/submissions", requestOptions{
		body:   quizBody,
		cookie: loginCookie.Value,
		csrf:   loginToken,
	})
	if quiz.Code != http.StatusOK {
		t.Fatalf("quiz submission: expected status 200, got %d: %s", quiz.Code, quiz.Body.String())
	}

	logout := authRequest(t, s.handler, http.MethodPost, "/auth/logout", requestOptions{
		cookie: loginCookie.Value,
		csrf:   loginToken,
	})
	if logout.Code != http.StatusOK {
		t.Fatalf("logout: expected status 200, got %d: %s", logout.Code, logout.Body.String())
	}

	output := logs.String()
	if output == "" {
		t.Fatal("expected the logging middleware to write a log line")
	}
	for _, secret := range []string{password, user.PasswordHash, sessionID, token, loginCookie.Value, loginToken} {
		if strings.Contains(output, secret) {
			t.Errorf("the log leaks a secret: %s", output)
		}
	}
}
