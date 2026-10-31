// Package service holds the business logic of the API; handlers stay thin.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// ContentService serves published topics and lessons.
type ContentService struct {
	repository repo.ContentRepository
}

// NewContentService creates a content service over repository.
func NewContentService(repository repo.ContentRepository) *ContentService {
	return &ContentService{repository: repository}
}

// ListTopics returns the published topics. The repository already filters on
// status; the service enforces the published-only rule again.
func (s *ContentService) ListTopics(ctx context.Context) ([]repo.Topic, error) {
	topics, err := s.repository.ListPublishedTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}

	published := make([]repo.Topic, 0, len(topics))
	for _, topic := range topics {
		if topic.Status == repo.StatusPublished {
			published = append(published, topic)
		}
	}
	return published, nil
}

// GetTopic returns a published topic with its published lessons.
func (s *ContentService) GetTopic(ctx context.Context, slug string) (repo.Topic, []repo.Lesson, error) {
	topic, err := s.repository.GetPublishedTopicBySlug(ctx, slug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Topic{}, nil, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Topic{}, nil, fmt.Errorf("get topic %q: %w", slug, err)
	}
	if topic.Status != repo.StatusPublished {
		return repo.Topic{}, nil, notFound("Topic not found.")
	}

	lessons, err := s.repository.ListPublishedLessonsByTopicID(ctx, topic.ID)
	if err != nil {
		return repo.Topic{}, nil, fmt.Errorf("list lessons of topic %q: %w", slug, err)
	}

	published := make([]repo.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		if lesson.Status == repo.StatusPublished {
			published = append(published, lesson)
		}
	}
	return topic, published, nil
}

// GetLesson returns a published lesson.
func (s *ContentService) GetLesson(ctx context.Context, id string) (repo.Lesson, error) {
	lesson, err := s.repository.GetPublishedLessonByID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get lesson %q: %w", id, err)
	}
	if lesson.Status != repo.StatusPublished {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	return lesson, nil
}

func notFound(message string) error {
	return &apperr.Error{Code: apperr.CodeNotFound, Message: message}
}
