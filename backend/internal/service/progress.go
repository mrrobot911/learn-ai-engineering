// Package service holds the business logic of the API; handlers stay thin.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// ProgressService handles lesson progress operations.
type ProgressService struct {
	progressRepo repo.ProgressRepository
	contentRepo  repo.ContentRepositoryForProgress
	now          func() time.Time
}

// NewProgressService creates a progress service over repositories.
func NewProgressService(progressRepo repo.ProgressRepository, contentRepo repo.ContentRepositoryForProgress) *ProgressService {
	return &ProgressService{
		progressRepo: progressRepo,
		contentRepo:  contentRepo,
		now:          time.Now,
	}
}

// MarkLessonComplete marks a lesson as complete for the user.
func (s *ProgressService) MarkLessonComplete(ctx context.Context, userID, lessonID string) ([]string, error) {
	// Verify lesson exists and is published
	exists, err := s.contentRepo.LessonExistsAndPublished(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("check lesson: %w", err)
	}
	if !exists {
		return nil, notFound("no published lesson with this id")
	}

	// Mark lesson as complete
	progress := repo.LessonProgress{
		UserID:      userID,
		LessonID:    lessonID,
		CompletedAt: s.now(),
	}
	if err := s.progressRepo.MarkLessonComplete(ctx, progress); err != nil {
		if err == repo.ErrAlreadyCompleted {
			// Idempotent: return current progress
			return s.progressRepo.GetUserProgress(ctx, userID)
		}
		return nil, fmt.Errorf("mark lesson complete: %w", err)
	}

	// Return updated progress
	return s.progressRepo.GetUserProgress(ctx, userID)
}

// UnmarkLessonComplete removes completion for a lesson.
func (s *ProgressService) UnmarkLessonComplete(ctx context.Context, userID, lessonID string) ([]string, error) {
	// Verify lesson exists and is published
	exists, err := s.contentRepo.LessonExistsAndPublished(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("check lesson: %w", err)
	}
	if !exists {
		return nil, notFound("no published lesson with this id")
	}

	// Remove completion
	if err := s.progressRepo.UnmarkLessonComplete(ctx, userID, lessonID); err != nil {
		return nil, fmt.Errorf("unmark lesson complete: %w", err)
	}

	// Return updated progress
	return s.progressRepo.GetUserProgress(ctx, userID)
}

// GetUserProgress returns the lesson IDs the user has marked complete.
func (s *ProgressService) GetUserProgress(ctx context.Context, userID string) ([]string, error) {
	return s.progressRepo.GetUserProgress(ctx, userID)
}
