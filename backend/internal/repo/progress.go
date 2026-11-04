// Package repo defines the progress data access layer. SQL belongs only here.
package repo

import (
	"context"
	"errors"
	"sync"
	"time"
)

// LessonProgress represents a user's completion status of a lesson.
type LessonProgress struct {
	UserID      string
	LessonID    string
	CompletedAt time.Time
}

// ErrAlreadyCompleted is returned when marking a lesson complete that is already complete.
var ErrAlreadyCompleted = errors.New("lesson already completed")

// ErrLessonNotFound is returned when a lesson does not exist or is not published.
var ErrLessonNotFound = errors.New("lesson not found")

// ProgressRepository stores lesson progress. SQL belongs only in the eventual
// database implementation.
type ProgressRepository interface {
	MarkLessonComplete(ctx context.Context, progress LessonProgress) error
	UnmarkLessonComplete(ctx context.Context, userID, lessonID string) error
	GetUserProgress(ctx context.Context, userID string) ([]string, error)
	LessonExistsAndPublished(ctx context.Context, lessonID string) (bool, error)
}

// MemoryProgressRepository is an in-memory ProgressRepository for unit tests and
// until a database repository replaces it.
type MemoryProgressRepository struct {
	mu          sync.RWMutex
	progressMap map[string][]LessonProgress // userID -> []LessonProgress
}

// NewMemoryProgressRepository creates an empty in-memory progress repository.
func NewMemoryProgressRepository() *MemoryProgressRepository {
	return &MemoryProgressRepository{
		progressMap: make(map[string][]LessonProgress),
	}
}

func (r *MemoryProgressRepository) MarkLessonComplete(ctx context.Context, progress LessonProgress) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already completed
	for _, p := range r.progressMap[progress.UserID] {
		if p.LessonID == progress.LessonID {
			return ErrAlreadyCompleted
		}
	}

	// Add the new progress
	r.progressMap[progress.UserID] = append(r.progressMap[progress.UserID], progress)
	return nil
}

func (r *MemoryProgressRepository) UnmarkLessonComplete(ctx context.Context, userID, lessonID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	progresses := r.progressMap[userID]
	for i, p := range progresses {
		if p.LessonID == lessonID {
			// Remove the progress
			r.progressMap[userID] = append(progresses[:i], progresses[i+1:]...)
			return nil
		}
	}

	// Lesson not marked, which is fine
	return nil
}

func (r *MemoryProgressRepository) GetUserProgress(ctx context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	progresses, ok := r.progressMap[userID]
	if !ok {
		return []string{}, nil
	}

	lessonIDs := make([]string, len(progresses))
	for i, p := range progresses {
		lessonIDs[i] = p.LessonID
	}
	return lessonIDs, nil
}

func (r *MemoryProgressRepository) LessonExistsAndPublished(ctx context.Context, lessonID string) (bool, error) {
	// For now, this is a placeholder. In the future, this would query the content repository
	// to check if the lesson exists and is published.
	// For the memory implementation, we'll return true for any lesson ID to keep tests simple.
	return true, nil
}
