package repo

import (
	"context"
	"sync"
)

// MemoryAuthRepository is an in-memory AuthRepository for unit tests and until
// a database repository replaces it.
type MemoryAuthRepository struct {
	mu          sync.RWMutex
	users       map[string]User
	usersByMail map[string]string
	sessions    map[string]Session
}

// NewMemoryAuthRepository creates an empty in-memory auth repository.
func NewMemoryAuthRepository() *MemoryAuthRepository {
	return &MemoryAuthRepository{
		users:       make(map[string]User),
		usersByMail: make(map[string]string),
		sessions:    make(map[string]Session),
	}
}

func (r *MemoryAuthRepository) CreateUser(_ context.Context, user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByMail[user.Email]; exists {
		return ErrConflict
	}
	if _, exists := r.users[user.ID]; exists {
		return ErrConflict
	}
	r.users[user.ID] = user
	r.usersByMail[user.Email] = user.ID
	return nil
}

func (r *MemoryAuthRepository) GetUserByID(_ context.Context, id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryAuthRepository) GetUserByEmail(_ context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.usersByMail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return r.users[id], nil
}

func (r *MemoryAuthRepository) CreateSession(_ context.Context, session Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; exists {
		return ErrConflict
	}
	r.sessions[session.ID] = session
	return nil
}

func (r *MemoryAuthRepository) GetSessionByID(_ context.Context, id string) (Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (r *MemoryAuthRepository) DeleteSession(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[id]; !ok {
		return ErrNotFound
	}
	delete(r.sessions, id)
	return nil
}

// UserCount reports the number of stored users. Tests use it to assert that a
// rejected registration created nothing.
func (r *MemoryAuthRepository) UserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}

// SessionCount reports the number of stored sessions. Tests use it to assert
// that a rejected request created nothing.
func (r *MemoryAuthRepository) SessionCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}
