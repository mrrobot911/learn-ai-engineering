package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

// unfilteredRepository returns every item it holds, including drafts, so the
// service's published-only rule is exercised on its own.
type unfilteredRepository struct {
	topics    []repo.Topic
	lessons   []repo.Lesson
	questions []repo.Question
	err       error
}

func (r unfilteredRepository) ListPublishedTopics(context.Context) ([]repo.Topic, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.topics, nil
}

func (r unfilteredRepository) GetPublishedTopicBySlug(_ context.Context, slug string) (repo.Topic, error) {
	if r.err != nil {
		return repo.Topic{}, r.err
	}
	for _, topic := range r.topics {
		if topic.Slug == slug {
			return topic, nil
		}
	}
	return repo.Topic{}, repo.ErrNotFound
}

func (r unfilteredRepository) ListPublishedLessonsByTopicID(_ context.Context, topicID string) ([]repo.Lesson, error) {
	if r.err != nil {
		return nil, r.err
	}
	lessons := make([]repo.Lesson, 0)
	for _, lesson := range r.lessons {
		if lesson.TopicID == topicID {
			lessons = append(lessons, lesson)
		}
	}
	return lessons, nil
}

func (r unfilteredRepository) GetPublishedLessonByID(_ context.Context, id string) (repo.Lesson, error) {
	if r.err != nil {
		return repo.Lesson{}, r.err
	}
	for _, lesson := range r.lessons {
		if lesson.ID == id {
			return lesson, nil
		}
	}
	return repo.Lesson{}, repo.ErrNotFound
}

// ListPublishedQuestionsByTopicID ignores the topic and the status on purpose,
// so the service's own filtering is exercised.
func (r unfilteredRepository) ListPublishedQuestionsByTopicID(context.Context, string) ([]repo.Question, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.questions, nil
}

func TestContentServiceListTopicsExcludesDrafts(t *testing.T) {
	s := NewContentService(unfilteredRepository{topics: []repo.Topic{
		{ID: "topic-1", Slug: "published", Status: repo.StatusPublished},
		{ID: "topic-2", Slug: "draft", Status: repo.StatusDraft},
	}})

	topics, err := s.ListTopics(context.Background())
	if err != nil {
		t.Fatalf("list topics: %v", err)
	}
	if len(topics) != 1 || topics[0].ID != "topic-1" {
		t.Errorf("expected only the published topic, got %v", topics)
	}
}

func TestContentServiceGetTopicExcludesDrafts(t *testing.T) {
	s := NewContentService(unfilteredRepository{
		topics: []repo.Topic{
			{ID: "topic-1", Slug: "published", Status: repo.StatusPublished},
			{ID: "topic-2", Slug: "draft", Status: repo.StatusDraft},
		},
		lessons: []repo.Lesson{
			{ID: "lesson-1", TopicID: "topic-1", Status: repo.StatusPublished},
			{ID: "lesson-2", TopicID: "topic-1", Status: repo.StatusDraft},
		},
	})

	topic, lessons, err := s.GetTopic(context.Background(), "published")
	if err != nil {
		t.Fatalf("get published topic: %v", err)
	}
	if topic.ID != "topic-1" {
		t.Errorf("expected topic-1, got %q", topic.ID)
	}
	if len(lessons) != 1 || lessons[0].ID != "lesson-1" {
		t.Errorf("expected only the published lesson, got %v", lessons)
	}

	_, _, err = s.GetTopic(context.Background(), "draft")
	if !hasCode(err, apperr.CodeNotFound) {
		t.Errorf("expected not_found for a draft topic, got %v", err)
	}
}

func TestContentServiceGetLessonExcludesDrafts(t *testing.T) {
	s := NewContentService(unfilteredRepository{lessons: []repo.Lesson{
		{ID: "lesson-1", Status: repo.StatusDraft},
	}})

	if _, err := s.GetLesson(context.Background(), "lesson-1"); !hasCode(err, apperr.CodeNotFound) {
		t.Errorf("expected not_found for a draft lesson, got %v", err)
	}
}

func TestContentServiceWrapsRepositoryError(t *testing.T) {
	cause := errors.New("query failed")
	s := NewContentService(unfilteredRepository{err: cause})

	_, err := s.ListTopics(context.Background())
	if !errors.Is(err, cause) {
		t.Errorf("expected the repository error to be wrapped, got %v", err)
	}
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		t.Errorf("expected an unexpected repository error to stay unmapped, got %v", err)
	}
}

func hasCode(err error, code apperr.Code) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Code == code
}
