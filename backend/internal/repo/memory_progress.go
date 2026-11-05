package repo

import (
	"context"
	"sync"
	"time"
)

type progressKey struct {
	userID   string
	lessonID string
}

// MemoryProgressRepository is an in-memory ProgressRepository for unit tests
// and until a database repository replaces it.
type MemoryProgressRepository struct {
	mu         sync.RWMutex
	completion map[progressKey]time.Time
}

// NewMemoryProgressRepository creates an empty in-memory progress repository.
func NewMemoryProgressRepository() *MemoryProgressRepository {
	return &MemoryProgressRepository{completion: make(map[progressKey]time.Time)}
}

func (r *MemoryProgressRepository) ListLessonProgress(_ context.Context, userID string) ([]LessonProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	progress := make([]LessonProgress, 0)
	for key, completedAt := range r.completion {
		if key.userID != userID {
			continue
		}
		progress = append(progress, LessonProgress{
			UserID:      key.userID,
			LessonID:    key.lessonID,
			CompletedAt: completedAt,
		})
	}
	return progress, nil
}

func (r *MemoryProgressRepository) MarkLessonComplete(_ context.Context, userID, lessonID string, completedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := progressKey{userID: userID, lessonID: lessonID}
	if _, exists := r.completion[key]; exists {
		return nil
	}
	r.completion[key] = completedAt
	return nil
}

func (r *MemoryProgressRepository) UnmarkLessonComplete(_ context.Context, userID, lessonID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.completion, progressKey{userID: userID, lessonID: lessonID})
	return nil
}
