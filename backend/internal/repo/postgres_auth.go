package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAuthRepository is a PostgreSQL implementation of AuthRepository.
type PostgresAuthRepository struct {
	pool *pgxpool.Pool
}

var _ AuthRepository = (*PostgresAuthRepository)(nil)

// NewPostgresAuthRepository creates an auth repository backed by pool.
func NewPostgresAuthRepository(pool *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{pool: pool}
}

func (r *PostgresAuthRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, is_maintainer, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.PasswordHash, user.IsMaintainer, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *PostgresAuthRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, is_maintainer, created_at
		 FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsMaintainer, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (r *PostgresAuthRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, is_maintainer, created_at
		 FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsMaintainer, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, session Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, csrf_token, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		session.ID, session.UserID, session.CSRFToken, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *PostgresAuthRepository) GetSessionByID(ctx context.Context, id string) (Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, csrf_token, created_at, expires_at
		 FROM sessions WHERE id = $1`, id).
		Scan(&s.ID, &s.UserID, &s.CSRFToken, &s.CreatedAt, &s.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	return s, err
}

func (r *PostgresAuthRepository) DeleteSession(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
