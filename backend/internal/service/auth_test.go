package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

func newTestAuthService(t *testing.T) (*AuthService, *repo.MemoryAuthRepository) {
	t.Helper()
	repository := repo.NewMemoryAuthRepository()
	return NewAuthService(repository), repository
}

func assertDomainError(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected an apperr.Error with code %q, got %v", code, err)
	}
	if domainErr.Code != code {
		t.Errorf("expected code %q, got %q", code, domainErr.Code)
	}
	if domainErr.Message == "" {
		t.Error("expected a non-empty message")
	}
}

func TestRegisterStoresSaltedArgon2idHashes(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()

	first, err := s.Register(ctx, "first@example.com", "password")
	if err != nil {
		t.Fatalf("register first user: %v", err)
	}
	second, err := s.Register(ctx, "second@example.com", "password")
	if err != nil {
		t.Fatalf("register second user: %v", err)
	}

	firstUser, err := repository.GetUserByEmail(ctx, "first@example.com")
	if err != nil {
		t.Fatalf("get first user: %v", err)
	}
	secondUser, err := repository.GetUserByEmail(ctx, "second@example.com")
	if err != nil {
		t.Fatalf("get second user: %v", err)
	}

	if !strings.HasPrefix(firstUser.PasswordHash, "$argon2id$") {
		t.Errorf("expected an Argon2id hash, got %q", firstUser.PasswordHash)
	}
	if strings.Contains(firstUser.PasswordHash, "password") {
		t.Errorf("stored hash contains the plaintext password: %q", firstUser.PasswordHash)
	}
	if firstUser.PasswordHash == secondUser.PasswordHash {
		t.Error("two users with the same password share a hash, so the salt is missing")
	}
	if first.User.PasswordHash != "" || second.User.PasswordHash != "" {
		t.Error("the returned user carries the password hash")
	}

	for _, tc := range []struct {
		hash     string
		password string
		want     bool
	}{
		{firstUser.PasswordHash, "password", true},
		{firstUser.PasswordHash, "password-wrong", false},
		{secondUser.PasswordHash, "password", true},
	} {
		matches, err := verifyPassword(tc.hash, tc.password)
		if err != nil {
			t.Fatalf("verify password: %v", err)
		}
		if matches != tc.want {
			t.Errorf("verify %q against stored hash: expected %v, got %v", tc.password, tc.want, matches)
		}
	}
}

func TestRegisterAcceptsPasswordBounds(t *testing.T) {
	for _, length := range []int{minPasswordLength, maxPasswordLength} {
		t.Run(strings.Repeat("x", length), func(t *testing.T) {
			s, _ := newTestAuthService(t)
			password := strings.Repeat("a", length)

			auth, err := s.Register(context.Background(), "user@example.com", password)
			if err != nil {
				t.Fatalf("register with a %d character password: %v", length, err)
			}
			if auth.Session.CSRFToken == "" {
				t.Error("expected a CSRF token")
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
		{"password too short", "user@example.com", strings.Repeat("a", minPasswordLength-1)},
		{"password too long", "user@example.com", strings.Repeat("a", maxPasswordLength+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repository := newTestAuthService(t)

			_, err := s.Register(context.Background(), tt.email, tt.password)
			assertDomainError(t, err, apperr.CodeValidation)
			if repository.UserCount() != 0 {
				t.Errorf("expected no user, got %d", repository.UserCount())
			}
			if repository.SessionCount() != 0 {
				t.Errorf("expected no session, got %d", repository.SessionCount())
			}
		})
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()

	if _, err := s.Register(ctx, "user@example.com", "password"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	before, err := repository.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("get stored user: %v", err)
	}
	users, sessions := repository.UserCount(), repository.SessionCount()

	_, err = s.Register(ctx, "user@example.com", "another-password")
	assertDomainError(t, err, apperr.CodeConflict)

	if got := repository.UserCount(); got != users {
		t.Errorf("expected %d users after the duplicate, got %d", users, got)
	}
	if got := repository.SessionCount(); got != sessions {
		t.Errorf("expected %d sessions after the duplicate, got %d", sessions, got)
	}
	after, err := repository.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("get stored user after duplicate: %v", err)
	}
	if after.PasswordHash != before.PasswordHash {
		t.Error("the duplicate registration changed the stored password hash")
	}
}

func TestLoginStartsASevenDaySession(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()
	fixed := time.Date(2026, time.November, 3, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fixed }

	if _, err := s.Register(ctx, "user@example.com", "password"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	auth, err := s.Login(ctx, "user@example.com", "password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if auth.User.Email != "user@example.com" {
		t.Errorf("expected the registered user, got %q", auth.User.Email)
	}
	if auth.User.PasswordHash != "" {
		t.Error("the login response carries the password hash")
	}
	if auth.Session.CSRFToken == "" {
		t.Error("expected a non-empty CSRF token")
	}
	if got := auth.Session.ExpiresAt.Sub(auth.Session.CreatedAt); got != SessionDuration {
		t.Errorf("expected a %s session, got %s", SessionDuration, got)
	}

	stored, err := repository.GetSessionByID(ctx, auth.Session.ID)
	if err != nil {
		t.Fatalf("get stored session: %v", err)
	}
	if !stored.ExpiresAt.Equal(fixed.Add(SessionDuration)) {
		t.Errorf("expected expires_at %s, got %s", fixed.Add(SessionDuration), stored.ExpiresAt)
	}
	if stored.CSRFToken != auth.Session.CSRFToken {
		t.Error("the returned CSRF token differs from the stored token")
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
		{"unknown email", "unknown@example.com", "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repository := newTestAuthService(t)
			ctx := context.Background()
			if _, err := s.Register(ctx, "user@example.com", "password"); err != nil {
				t.Fatalf("register user: %v", err)
			}
			sessions := repository.SessionCount()

			_, err := s.Login(ctx, tt.email, tt.password)
			assertDomainError(t, err, apperr.CodeAuthentication)
			if got := repository.SessionCount(); got != sessions {
				t.Errorf("expected %d sessions after the failed login, got %d", sessions, got)
			}
		})
	}
}

func TestLoginRejectsMissingOrInvalidInput(t *testing.T) {
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
			s, repository := newTestAuthService(t)
			ctx := context.Background()
			if _, err := s.Register(ctx, "user@example.com", "password"); err != nil {
				t.Fatalf("register user: %v", err)
			}
			sessions := repository.SessionCount()

			_, err := s.Login(ctx, tt.email, tt.password)
			assertDomainError(t, err, apperr.CodeValidation)
			if got := repository.SessionCount(); got != sessions {
				t.Errorf("expected %d sessions after the invalid login, got %d", sessions, got)
			}
		})
	}
}

func TestCurrentSession(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()
	fixed := time.Date(2026, time.November, 3, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fixed }

	registered, err := s.Register(ctx, "user@example.com", "password")
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	t.Run("valid", func(t *testing.T) {
		auth, err := s.CurrentSession(ctx, registered.Session.ID)
		if err != nil {
			t.Fatalf("current session: %v", err)
		}
		if auth.User.ID != registered.User.ID || auth.User.Email != registered.User.Email {
			t.Errorf("expected the registered user, got %v", auth.User)
		}
		if auth.Session.CSRFToken != registered.Session.CSRFToken {
			t.Error("expected the stored CSRF token")
		}
	})

	for _, tt := range []struct {
		name      string
		sessionID string
	}{
		{"missing", ""},
		{"unknown", "unknown-session"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.CurrentSession(ctx, tt.sessionID)
			assertDomainError(t, err, apperr.CodeAuthentication)
		})
	}

	t.Run("expired", func(t *testing.T) {
		expired := repo.Session{
			ID:        "expired-session",
			UserID:    registered.User.ID,
			CSRFToken: "expired-csrf",
			CreatedAt: fixed.Add(-SessionDuration),
			ExpiresAt: fixed,
		}
		if err := repository.CreateSession(ctx, expired); err != nil {
			t.Fatalf("create expired session: %v", err)
		}

		_, err := s.CurrentSession(ctx, expired.ID)
		assertDomainError(t, err, apperr.CodeAuthentication)
	})
}

func TestLogoutEndsTheSession(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()
	auth, err := s.Register(ctx, "user@example.com", "password")
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	if err := s.Logout(ctx, auth.Session.ID); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if got := repository.SessionCount(); got != 0 {
		t.Errorf("expected no sessions after logout, got %d", got)
	}
	if _, err := repository.GetSessionByID(ctx, auth.Session.ID); !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("expected the session to be deleted, got %v", err)
	}
}

func TestLogoutRejectsInvalidSessions(t *testing.T) {
	s, repository := newTestAuthService(t)
	ctx := context.Background()
	fixed := time.Date(2026, time.November, 3, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fixed }

	if _, err := s.Register(ctx, "user@example.com", "password"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	before := repository.SessionCount()

	for _, tt := range []struct {
		name      string
		sessionID string
	}{
		{"missing", ""},
		{"unknown", "unknown-session"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Logout(ctx, tt.sessionID)
			assertDomainError(t, err, apperr.CodeAuthentication)
		})
	}

	t.Run("expired", func(t *testing.T) {
		user, err := repository.GetUserByEmail(ctx, "user@example.com")
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		expired := repo.Session{
			ID:        "expired-session",
			UserID:    user.ID,
			CSRFToken: "expired-csrf",
			CreatedAt: fixed.Add(-SessionDuration),
			ExpiresAt: fixed,
		}
		if err := repository.CreateSession(ctx, expired); err != nil {
			t.Fatalf("create expired session: %v", err)
		}
		before++

		err = s.Logout(ctx, expired.ID)
		assertDomainError(t, err, apperr.CodeAuthentication)
		if got := repository.SessionCount(); got != before {
			t.Errorf("expected the expired session to remain for cleanup, got %d sessions", got)
		}
	})

	if got := repository.SessionCount(); got != before {
		t.Errorf("expected %d sessions after the rejected logouts, got %d", before, got)
	}
}
