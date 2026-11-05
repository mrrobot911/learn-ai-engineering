package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// ProgressService reads and changes the lesson progress of a user.
type ProgressService struct {
	progress repo.ProgressRepository
	content  repo.ContentRepository
	now      func() time.Time
}

// NewProgressService creates a progress service over the progress and content
// repositories.
func NewProgressService(progress repo.ProgressRepository, content repo.ContentRepository) *ProgressService {
	return NewProgressServiceWithClock(progress, content, time.Now)
}

// NewProgressServiceWithClock creates a progress service that reads the clock
// from now, so tests can pin the completion time.
func NewProgressServiceWithClock(progress repo.ProgressRepository, content repo.ContentRepository, now func() time.Time) *ProgressService {
	return &ProgressService{progress: progress, content: content, now: now}
}

// ListCompletedLessonIDs returns the ids of the lessons the user has marked
// complete. The result is never nil.
func (s *ProgressService) ListCompletedLessonIDs(ctx context.Context, userID string) ([]string, error) {
	progress, err := s.progress.ListLessonProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list lesson progress of user %q: %w", userID, err)
	}

	ids := make([]string, 0, len(progress))
	for _, record := range progress {
		ids = append(ids, record.LessonID)
	}
	return ids, nil
}

// MarkLessonComplete marks a published lesson complete for the user and
// returns the user's completed lesson ids. Marking it again keeps the original
// completion time.
func (s *ProgressService) MarkLessonComplete(ctx context.Context, userID, lessonID string) ([]string, error) {
	if err := s.requirePublishedLesson(ctx, lessonID); err != nil {
		return nil, err
	}
	if err := s.progress.MarkLessonComplete(ctx, userID, lessonID, s.now()); err != nil {
		return nil, fmt.Errorf("mark lesson %q complete for user %q: %w", lessonID, userID, err)
	}
	return s.ListCompletedLessonIDs(ctx, userID)
}

// UnmarkLessonComplete removes the completion of a published lesson for the
// user and returns the user's completed lesson ids. Undoing an unmarked lesson
// is not an error.
func (s *ProgressService) UnmarkLessonComplete(ctx context.Context, userID, lessonID string) ([]string, error) {
	if err := s.requirePublishedLesson(ctx, lessonID); err != nil {
		return nil, err
	}
	if err := s.progress.UnmarkLessonComplete(ctx, userID, lessonID); err != nil {
		return nil, fmt.Errorf("unmark lesson %q complete for user %q: %w", lessonID, userID, err)
	}
	return s.ListCompletedLessonIDs(ctx, userID)
}

// requirePublishedLesson rejects an unknown or unpublished lesson with
// not_found before any progress record is written.
func (s *ProgressService) requirePublishedLesson(ctx context.Context, lessonID string) error {
	lesson, err := s.content.GetPublishedLessonByID(ctx, lessonID)
	if errors.Is(err, repo.ErrNotFound) {
		return notFound("Lesson not found.")
	}
	if err != nil {
		return fmt.Errorf("get lesson %q: %w", lessonID, err)
	}
	if lesson.Status != repo.StatusPublished {
		return notFound("Lesson not found.")
	}
	return nil
}
