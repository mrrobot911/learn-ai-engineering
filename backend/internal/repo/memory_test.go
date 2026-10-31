package repo

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryContentRepositoryListsPublishedTopicsInPositionOrder(t *testing.T) {
	r := NewMemoryContentRepository([]Topic{
		{ID: "topic-1", Slug: "first", Position: 1, Status: StatusPublished},
		{ID: "topic-2", Slug: "draft", Position: 2, Status: StatusDraft},
		{ID: "topic-3", Slug: "third", Position: 3, Status: StatusPublished},
	}, nil)

	topics, err := r.ListPublishedTopics(context.Background())
	if err != nil {
		t.Fatalf("list published topics: %v", err)
	}

	if len(topics) != 2 {
		t.Fatalf("expected 2 published topics, got %d: %v", len(topics), topics)
	}
	if topics[0].ID != "topic-1" || topics[1].ID != "topic-3" {
		t.Errorf("expected topics in position order with the draft excluded, got %v", topics)
	}
}

func TestMemoryContentRepositoryReturnsEmptyTopicList(t *testing.T) {
	r := NewMemoryContentRepository([]Topic{
		{ID: "topic-1", Slug: "draft", Position: 1, Status: StatusDraft},
	}, nil)

	topics, err := r.ListPublishedTopics(context.Background())
	if err != nil {
		t.Fatalf("list published topics: %v", err)
	}
	if topics == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(topics) != 0 {
		t.Errorf("expected no topics, got %v", topics)
	}
}

func TestMemoryContentRepositoryGetPublishedTopicBySlug(t *testing.T) {
	r := NewMemoryContentRepository([]Topic{
		{ID: "topic-1", Slug: "published", Position: 1, Status: StatusPublished},
		{ID: "topic-2", Slug: "draft", Position: 2, Status: StatusDraft},
	}, nil)

	t.Run("published", func(t *testing.T) {
		topic, err := r.GetPublishedTopicBySlug(context.Background(), "published")
		if err != nil {
			t.Fatalf("get published topic: %v", err)
		}
		if topic.ID != "topic-1" {
			t.Errorf("expected topic-1, got %q", topic.ID)
		}
	})

	for _, slug := range []string{"unknown", "draft"} {
		t.Run(slug, func(t *testing.T) {
			if _, err := r.GetPublishedTopicBySlug(context.Background(), slug); !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

func TestMemoryContentRepositoryListsPublishedLessonsInPositionOrder(t *testing.T) {
	r := NewMemoryContentRepository(nil, []Lesson{
		{ID: "lesson-2", TopicID: "topic-1", Position: 2, Status: StatusPublished},
		{ID: "lesson-1", TopicID: "topic-1", Position: 1, Status: StatusPublished},
		{ID: "lesson-draft", TopicID: "topic-1", Position: 3, Status: StatusDraft},
		{ID: "lesson-other", TopicID: "topic-2", Position: 1, Status: StatusPublished},
	})

	lessons, err := r.ListPublishedLessonsByTopicID(context.Background(), "topic-1")
	if err != nil {
		t.Fatalf("list published lessons: %v", err)
	}

	if len(lessons) != 2 {
		t.Fatalf("expected 2 published lessons, got %d: %v", len(lessons), lessons)
	}
	if lessons[0].ID != "lesson-1" || lessons[1].ID != "lesson-2" {
		t.Errorf("expected lessons in position order with the draft excluded, got %v", lessons)
	}
}

func TestMemoryContentRepositoryReturnsEmptyLessonList(t *testing.T) {
	r := NewMemoryContentRepository(nil, nil)

	lessons, err := r.ListPublishedLessonsByTopicID(context.Background(), "topic-1")
	if err != nil {
		t.Fatalf("list published lessons: %v", err)
	}
	if lessons == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(lessons) != 0 {
		t.Errorf("expected no lessons, got %v", lessons)
	}
}

func TestMemoryContentRepositoryGetPublishedLessonByID(t *testing.T) {
	r := NewMemoryContentRepository(nil, []Lesson{
		{ID: "lesson-1", TopicID: "topic-1", Status: StatusPublished},
		{ID: "lesson-draft", TopicID: "topic-1", Status: StatusDraft},
	})

	t.Run("published", func(t *testing.T) {
		lesson, err := r.GetPublishedLessonByID(context.Background(), "lesson-1")
		if err != nil {
			t.Fatalf("get published lesson: %v", err)
		}
		if lesson.ID != "lesson-1" {
			t.Errorf("expected lesson-1, got %q", lesson.ID)
		}
	})

	for _, id := range []string{"unknown", "lesson-draft"} {
		t.Run(id, func(t *testing.T) {
			if _, err := r.GetPublishedLessonByID(context.Background(), id); !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}
		})
	}
}
