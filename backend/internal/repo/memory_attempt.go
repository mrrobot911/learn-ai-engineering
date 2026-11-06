package repo

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryAttemptRepository is an in-memory AttemptRepository for unit tests and
// until a database repository replaces it.
type MemoryAttemptRepository struct {
	mu       sync.RWMutex
	nextID   int
	attempts []QuizAttempt
}

// NewMemoryAttemptRepository creates an empty in-memory attempt repository.
func NewMemoryAttemptRepository() *MemoryAttemptRepository {
	return &MemoryAttemptRepository{}
}

func (r *MemoryAttemptRepository) CreateAttempt(_ context.Context, attempt QuizAttempt) (QuizAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	attempt.ID = fmt.Sprintf("attempt-%d", r.nextID)
	attempt.Answers = cloneAttemptAnswers(attempt.Answers)
	r.attempts = append(r.attempts, attempt)
	return cloneAttempt(attempt), nil
}

func (r *MemoryAttemptRepository) ListAttemptsByUserID(_ context.Context, userID string) ([]QuizAttempt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	attempts := make([]QuizAttempt, 0)
	for _, attempt := range r.attempts {
		if attempt.UserID == userID {
			attempts = append(attempts, cloneAttempt(attempt))
		}
	}
	sort.SliceStable(attempts, func(i, j int) bool {
		return attempts[i].CreatedAt.After(attempts[j].CreatedAt)
	})
	return attempts, nil
}

func (r *MemoryAttemptRepository) GetAttemptByID(_ context.Context, id string) (QuizAttempt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, attempt := range r.attempts {
		if attempt.ID == id {
			return cloneAttempt(attempt), nil
		}
	}
	return QuizAttempt{}, ErrAttemptNotFound
}

func cloneAttempt(attempt QuizAttempt) QuizAttempt {
	attempt.Answers = cloneAttemptAnswers(attempt.Answers)
	return attempt
}

func cloneAttemptAnswers(answers []QuizAttemptAnswer) []QuizAttemptAnswer {
	if answers == nil {
		return nil
	}
	return append([]QuizAttemptAnswer(nil), answers...)
}
