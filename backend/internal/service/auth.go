package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	pwhash "github.com/mrrobot911/learn-ai-engineering/internal/password"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128

	// SessionDuration is the fixed lifetime of a session and its cookie.
	SessionDuration = 7 * 24 * time.Hour
)

// AuthSession is an authenticated session together with its user.
type AuthSession struct {
	User    repo.User
	Session repo.Session
}

// AuthService registers users, authenticates them and manages their sessions.
type AuthService struct {
	repository repo.AuthRepository
	now        func() time.Time
	random     io.Reader
}

// NewAuthService creates an auth service over repository.
func NewAuthService(repository repo.AuthRepository) *AuthService {
	return &AuthService{
		repository: repository,
		now:        time.Now,
		random:     rand.Reader,
	}
}

// Register creates a user with a salted Argon2id password hash and starts a
// session for it.
func (s *AuthService) Register(ctx context.Context, email, password string) (AuthSession, error) {
	if err := validateEmail(email); err != nil {
		return AuthSession{}, err
	}
	if err := validatePassword(password); err != nil {
		return AuthSession{}, err
	}

	passwordHash, err := pwhash.Hash(password)
	if err != nil {
		return AuthSession{}, fmt.Errorf("hash password for new user: %w", err)
	}

	now := s.now()
	user := repo.User{
		ID:           "",
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}
	user.ID, err = s.newToken()
	if err != nil {
		return AuthSession{}, fmt.Errorf("generate user id: %w", err)
	}

	if err := s.repository.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repo.ErrConflict) {
			return AuthSession{}, &apperr.Error{
				Code:    apperr.CodeConflict,
				Message: "An account with this email already exists.",
			}
		}
		return AuthSession{}, fmt.Errorf("create user: %w", err)
	}

	return s.startSession(ctx, user, now)
}

// Login checks the credentials and starts a session. A wrong password and an
// unknown email answer the same authentication error.
func (s *AuthService) Login(ctx context.Context, email, password string) (AuthSession, error) {
	if err := validateEmail(email); err != nil {
		return AuthSession{}, err
	}
	if password == "" {
		return AuthSession{}, validationError("The password is required.")
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if errors.Is(err, repo.ErrNotFound) {
		return AuthSession{}, invalidCredentials()
	}
	if err != nil {
		return AuthSession{}, fmt.Errorf("get user by email: %w", err)
	}

	matches, err := pwhash.Verify(user.PasswordHash, password)
	if err != nil {
		return AuthSession{}, fmt.Errorf("verify password: %w", err)
	}
	if !matches {
		return AuthSession{}, invalidCredentials()
	}

	return s.startSession(ctx, user, s.now())
}

// CurrentSession returns the session and user for a session id. A missing,
// unknown or expired session is an authentication error.
func (s *AuthService) CurrentSession(ctx context.Context, sessionID string) (AuthSession, error) {
	if sessionID == "" {
		return AuthSession{}, authenticationError("Authentication is required.")
	}

	session, err := s.repository.GetSessionByID(ctx, sessionID)
	if errors.Is(err, repo.ErrNotFound) {
		return AuthSession{}, authenticationError("The session is invalid or has expired.")
	}
	if err != nil {
		return AuthSession{}, fmt.Errorf("get session: %w", err)
	}
	if !s.now().Before(session.ExpiresAt) {
		return AuthSession{}, authenticationError("The session is invalid or has expired.")
	}

	user, err := s.repository.GetUserByID(ctx, session.UserID)
	if errors.Is(err, repo.ErrNotFound) {
		return AuthSession{}, authenticationError("The session is invalid or has expired.")
	}
	if err != nil {
		return AuthSession{}, fmt.Errorf("get session user: %w", err)
	}

	return newAuthSession(user, session), nil
}

// newAuthSession keeps the stored password hash inside the service; handlers
// only ever see the public fields of the user.
func newAuthSession(user repo.User, session repo.Session) AuthSession {
	user.PasswordHash = ""
	return AuthSession{User: user, Session: session}
}

// Logout ends a valid session.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return authenticationError("Authentication is required.")
	}

	session, err := s.repository.GetSessionByID(ctx, sessionID)
	if errors.Is(err, repo.ErrNotFound) {
		return authenticationError("The session is invalid or has expired.")
	}
	if err != nil {
		return fmt.Errorf("get session for logout: %w", err)
	}
	if !s.now().Before(session.ExpiresAt) {
		return authenticationError("The session is invalid or has expired.")
	}

	if err := s.repository.DeleteSession(ctx, sessionID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return authenticationError("The session is invalid or has expired.")
		}
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *AuthService) startSession(ctx context.Context, user repo.User, now time.Time) (AuthSession, error) {
	sessionID, err := s.newToken()
	if err != nil {
		return AuthSession{}, fmt.Errorf("generate session id: %w", err)
	}
	csrfToken, err := s.newToken()
	if err != nil {
		return AuthSession{}, fmt.Errorf("generate csrf token: %w", err)
	}

	session := repo.Session{
		ID:        sessionID,
		UserID:    user.ID,
		CSRFToken: csrfToken,
		CreatedAt: now,
		ExpiresAt: now.Add(SessionDuration),
	}
	if err := s.repository.CreateSession(ctx, session); err != nil {
		return AuthSession{}, fmt.Errorf("create session: %w", err)
	}

	return newAuthSession(user, session), nil
}

func (s *AuthService) newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(s.random, raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func validateEmail(email string) error {
	if email == "" {
		return validationError("A valid email address is required.")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || strings.TrimSpace(email) != email {
		return validationError("A valid email address is required.")
	}
	return nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength || length > maxPasswordLength {
		return validationError(fmt.Sprintf("The password must be between %d and %d characters.", minPasswordLength, maxPasswordLength))
	}
	return nil
}

func invalidCredentials() error {
	return authenticationError("Email or password is incorrect.")
}

func authenticationError(message string) error {
	return &apperr.Error{Code: apperr.CodeAuthentication, Message: message}
}
