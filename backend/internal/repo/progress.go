package repo

import (
	"context"
	"time"
)

// LessonProgress is the completion of one lesson by one user. A record only
// exists for a completed lesson.
type LessonProgress struct {
	UserID      string
	LessonID    string
	CompletedAt time.Time
}

// ProgressRepository stores lesson completions per user. Marking an already
// completed lesson keeps its completion time; undoing an unmarked lesson is
// not an error.
type ProgressRepository interface {
	ListLessonProgress(ctx context.Context, userID string) ([]LessonProgress, error)
	MarkLessonComplete(ctx context.Context, userID, lessonID string, completedAt time.Time) error
	UnmarkLessonComplete(ctx context.Context, userID, lessonID string) error
}
