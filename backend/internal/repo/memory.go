package repo

import (
	"context"
	"sort"
)

// MemoryContentRepository is an in-memory ContentRepository for unit tests and
// until a database repository replaces it.
type MemoryContentRepository struct {
	topics    []Topic
	lessons   []Lesson
	questions []Question
}

// NewMemoryContentRepository creates a repository over copies of topics,
// lessons and questions.
func NewMemoryContentRepository(topics []Topic, lessons []Lesson, questions []Question) *MemoryContentRepository {
	return &MemoryContentRepository{
		topics:    append([]Topic(nil), topics...),
		lessons:   append([]Lesson(nil), lessons...),
		questions: append([]Question(nil), questions...),
	}
}

func (r *MemoryContentRepository) ListPublishedTopics(_ context.Context) ([]Topic, error) {
	topics := make([]Topic, 0, len(r.topics))
	for _, topic := range r.topics {
		if topic.Status == StatusPublished {
			topics = append(topics, topic)
		}
	}
	sort.SliceStable(topics, func(i, j int) bool {
		return topics[i].Position < topics[j].Position
	})
	return topics, nil
}

func (r *MemoryContentRepository) GetPublishedTopicBySlug(_ context.Context, slug string) (Topic, error) {
	for _, topic := range r.topics {
		if topic.Slug == slug && topic.Status == StatusPublished {
			return topic, nil
		}
	}
	return Topic{}, ErrNotFound
}

func (r *MemoryContentRepository) ListPublishedLessonsByTopicID(_ context.Context, topicID string) ([]Lesson, error) {
	lessons := make([]Lesson, 0, len(r.lessons))
	for _, lesson := range r.lessons {
		if lesson.TopicID == topicID && lesson.Status == StatusPublished {
			lessons = append(lessons, lesson)
		}
	}
	sort.SliceStable(lessons, func(i, j int) bool {
		return lessons[i].Position < lessons[j].Position
	})
	return lessons, nil
}

func (r *MemoryContentRepository) GetPublishedLessonByID(_ context.Context, id string) (Lesson, error) {
	for _, lesson := range r.lessons {
		if lesson.ID == id && lesson.Status == StatusPublished {
			return lesson, nil
		}
	}
	return Lesson{}, ErrNotFound
}

func (r *MemoryContentRepository) ListPublishedQuestionsByTopicID(_ context.Context, topicID string) ([]Question, error) {
	questions := make([]Question, 0, len(r.questions))
	for _, question := range r.questions {
		if question.TopicID != topicID || question.Status != StatusPublished {
			continue
		}
		options := append([]Option(nil), question.Options...)
		sort.SliceStable(options, func(i, j int) bool {
			return options[i].Position < options[j].Position
		})
		question.Options = options
		questions = append(questions, question)
	}
	return questions, nil
}
