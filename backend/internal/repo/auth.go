package repo

import (
	"context"
	"errors"
	"time"
)

// ErrConflict is returned when a record would duplicate an existing one, for
// example a user with an email that is already registered.
var ErrConflict = errors.New("record already exists")

// User is an account of the platform. PasswordHash holds the encoded salted
// Argon2id hash, never the password.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	IsMaintainer bool
	CreatedAt    time.Time
}

// Session is an authenticated session. CSRFToken stays valid until the session
// ends.
type Session struct {
	ID        string
	UserID    string
	CSRFToken string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// AuthRepository stores users and sessions. SQL belongs only in the eventual
// database implementation.
type AuthRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateSession(ctx context.Context, session Session) error
	GetSessionByID(ctx context.Context, id string) (Session, error)
	DeleteSession(ctx context.Context, id string) error
}
