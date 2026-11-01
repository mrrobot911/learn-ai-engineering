// Package service holds the business logic of the API; handlers stay thin.
package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// quizSize is the number of questions a quiz returns when the topic has enough
// published questions. Version 1 does not let the client choose it.
const quizSize = 5

// ContentService serves published topics and lessons and starts quizzes.
type ContentService struct {
	repository repo.ContentRepository
	random     *rand.Rand
	randomMu   sync.Mutex
}

// NewContentService creates a content service over repository.
func NewContentService(repository repo.ContentRepository) *ContentService {
	return NewContentServiceWithRandom(repository, rand.New(rand.NewSource(time.Now().UnixNano())))
}

// NewContentServiceWithRandom creates a content service that draws its quiz
// questions from random, so tests can pin the sequence.
func NewContentServiceWithRandom(repository repo.ContentRepository, random *rand.Rand) *ContentService {
	return &ContentService{repository: repository, random: random}
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
	topic, err := s.getPublishedTopic(ctx, slug)
	if err != nil {
		return repo.Topic{}, nil, err
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

// StartQuiz returns the questions of a quiz on the topic: quizSize published
// questions drawn at random without replacement, or all of them when the topic
// has fewer. The options keep their stored position order.
func (s *ContentService) StartQuiz(ctx context.Context, slug string) ([]repo.Question, error) {
	topic, err := s.getPublishedTopic(ctx, slug)
	if err != nil {
		return nil, err
	}

	questions, err := s.repository.ListPublishedQuestionsByTopicID(ctx, topic.ID)
	if err != nil {
		return nil, fmt.Errorf("list questions of topic %q: %w", slug, err)
	}

	// The repository already filters on status and topic; the service enforces
	// the published-only rule again.
	published := make([]repo.Question, 0, len(questions))
	for _, question := range questions {
		if question.TopicID == topic.ID && question.Status == repo.StatusPublished {
			published = append(published, question)
		}
	}

	s.randomMu.Lock()
	s.random.Shuffle(len(published), func(i, j int) {
		published[i], published[j] = published[j], published[i]
	})
	s.randomMu.Unlock()

	if len(published) > quizSize {
		published = published[:quizSize]
	}
	return published, nil
}

func (s *ContentService) getPublishedTopic(ctx context.Context, slug string) (repo.Topic, error) {
	topic, err := s.repository.GetPublishedTopicBySlug(ctx, slug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Topic{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Topic{}, fmt.Errorf("get topic %q: %w", slug, err)
	}
	if topic.Status != repo.StatusPublished {
		return repo.Topic{}, notFound("Topic not found.")
	}
	return topic, nil
}

func notFound(message string) error {
	return &apperr.Error{Code: apperr.CodeNotFound, Message: message}
}
