package service

import (
	"context"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// WithKey is the context key for the authenticated session.
type contextKey struct{}

var WithKey = contextKey{}

func TestMarkLessonComplete(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
		{ID: "lesson2", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	service := NewProgressService(progressRepo, contentRepo)

	tests := []struct {
		name         string
		userID       string
		lessonID     string
		wantProgress []string
		wantErr      bool
		errCode      apperr.Code
	}{
		{
			name:         "mark lesson complete successfully",
			userID:       "user1",
			lessonID:     "lesson1",
			wantProgress: []string{"lesson1"},
			wantErr:      false,
		},
		{
			name:         "mark lesson complete again (idempotent)",
			userID:       "user1",
			lessonID:     "lesson1",
			wantProgress: []string{"lesson1"},
			wantErr:      false,
		},
		{
			name:         "mark different lesson",
			userID:       "user1",
			lessonID:     "lesson2",
			wantProgress: []string{"lesson1", "lesson2"},
			wantErr:      false,
		},
		{
			name:     "mark non-existent lesson",
			userID:   "user1",
			lessonID: "nonexistent",
			wantErr:  true,
			errCode:  apperr.CodeNotFound,
		},
		{
			name:     "mark draft lesson",
			userID:   "user1",
			lessonID: "draft-lesson",
			wantErr:  true,
			errCode:  apperr.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress, err := service.MarkLessonComplete(context.Background(), tt.userID, tt.lessonID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkLessonComplete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr && err.(*apperr.Error).Code != tt.errCode {
				t.Errorf("MarkLessonComplete() error code = %v, want %v", err.(*apperr.Error).Code, tt.errCode)
			}
			if !tt.wantErr {
				if len(progress) != len(tt.wantProgress) {
					t.Errorf("MarkLessonComplete() progress length = %v, want %v", len(progress), len(tt.wantProgress))
				}
				for i, id := range tt.wantProgress {
					if i >= len(progress) || progress[i] != id {
						t.Errorf("MarkLessonComplete() progress[%d] = %v, want %v", i, progress[i], id)
					}
				}
			}
		})
	}
}

func TestUnmarkLessonComplete(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
		{ID: "lesson2", TopicID: "topic1", Status: repo.StatusPublished},
		{ID: "lesson3", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	service := NewProgressService(progressRepo, contentRepo)

	// First, mark some lessons
	service.MarkLessonComplete(context.Background(), "user1", "lesson1")
	service.MarkLessonComplete(context.Background(), "user1", "lesson2")

	tests := []struct {
		name         string
		userID       string
		lessonID     string
		wantProgress []string
		wantErr      bool
	}{
		{
			name:         "unmark existing lesson",
			userID:       "user1",
			lessonID:     "lesson1",
			wantProgress: []string{"lesson2"},
			wantErr:      false,
		},
		{
			name:         "unmark non-marked lesson",
			userID:       "user1",
			lessonID:     "lesson3",
			wantProgress: []string{"lesson2"},
			wantErr:      false,
		},
		{
			name:         "unmark last lesson",
			userID:       "user1",
			lessonID:     "lesson2",
			wantProgress: []string{},
			wantErr:      false,
		},
		{
			name:     "unmark non-existent lesson",
			userID:   "user1",
			lessonID: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress, err := service.UnmarkLessonComplete(context.Background(), tt.userID, tt.lessonID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarkLessonComplete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				if err.(*apperr.Error).Code != apperr.CodeNotFound {
					t.Errorf("UnmarkLessonComplete() error code = %v, want %v", err.(*apperr.Error).Code, apperr.CodeNotFound)
				}
			}
			if !tt.wantErr {
				if len(progress) != len(tt.wantProgress) {
					t.Errorf("UnmarkLessonComplete() progress length = %v, want %v", len(progress), len(tt.wantProgress))
				}
				for i, id := range tt.wantProgress {
					if i >= len(progress) || progress[i] != id {
						t.Errorf("UnmarkLessonComplete() progress[%d] = %v, want %v", i, progress[i], id)
					}
				}
			}
		})
	}
}

func TestGetUserProgress(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	service := NewProgressService(progressRepo, contentRepo)

	// Get progress for user without any completed lessons
	progress, err := service.GetUserProgress(context.Background(), "user1")
	if err != nil {
		t.Errorf("GetUserProgress() error = %v", err)
	}
	if len(progress) != 0 {
		t.Errorf("GetUserProgress() progress = %v, want []", progress)
	}

	// Mark some lessons
	service.MarkLessonComplete(context.Background(), "user1", "lesson1")
	service.MarkLessonComplete(context.Background(), "user2", "lesson1")

	// Get progress for user1
	progress, err = service.GetUserProgress(context.Background(), "user1")
	if err != nil {
		t.Errorf("GetUserProgress() error = %v", err)
	}
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("GetUserProgress() progress = %v, want [lesson1]", progress)
	}

	// Get progress for user2
	progress, err = service.GetUserProgress(context.Background(), "user2")
	if err != nil {
		t.Errorf("GetUserProgress() error = %v", err)
	}
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("GetUserProgress() progress = %v, want [lesson1]", progress)
	}
}

func TestMarkLessonCompleteIdempotent(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	service := NewProgressService(progressRepo, contentRepo)

	// Mark the same lesson twice
	progress1, err := service.MarkLessonComplete(context.Background(), "user1", "lesson1")
	if err != nil {
		t.Errorf("First MarkLessonComplete() error = %v", err)
	}
	if len(progress1) != 1 || progress1[0] != "lesson1" {
		t.Errorf("First MarkLessonComplete() progress = %v, want [lesson1]", progress1)
	}

	progress2, err := service.MarkLessonComplete(context.Background(), "user1", "lesson1")
	if err != nil {
		t.Errorf("Second MarkLessonComplete() error = %v", err)
	}
	if len(progress2) != 1 || progress2[0] != "lesson1" {
		t.Errorf("Second MarkLessonComplete() progress = %v, want [lesson1]", progress2)
	}

	// Both should be identical
	for i := range progress1 {
		if progress1[i] != progress2[i] {
			t.Errorf("Marking lesson twice should be idempotent, got %v and %v", progress1, progress2)
		}
	}
}

func TestUnmarkLessonCompleteSafeToRepeat(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	service := NewProgressService(progressRepo, contentRepo)

	// Try to unmark a lesson that is not marked
	progress1, err := service.UnmarkLessonComplete(context.Background(), "user1", "lesson1")
	if err != nil {
		t.Errorf("First UnmarkLessonComplete() error = %v", err)
	}
	if len(progress1) != 0 {
		t.Errorf("First UnmarkLessonComplete() progress = %v, want []", progress1)
	}

	// Mark the lesson
	service.MarkLessonComplete(context.Background(), "user1", "lesson1")

	// Unmark it
	progress2, err := service.UnmarkLessonComplete(context.Background(), "user1", "lesson1")
	if err != nil {
		t.Errorf("Second UnmarkLessonComplete() error = %v", err)
	}
	if len(progress2) != 0 {
		t.Errorf("Second UnmarkLessonComplete() progress = %v, want []", progress2)
	}

	// Unmark again (should be safe)
	progress3, err := service.UnmarkLessonComplete(context.Background(), "user1", "lesson1")
	if err != nil {
		t.Errorf("Third UnmarkLessonComplete() error = %v", err)
	}
	if len(progress3) != 0 {
		t.Errorf("Third UnmarkLessonComplete() progress = %v, want []", progress3)
	}
}
