package repo

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryContentRepository is an in-memory ContentRepository for unit tests and
// until a database repository replaces it.
type MemoryContentRepository struct {
	mu          sync.RWMutex
	nextLesson  int
	nextQuest   int
	nextOption  int
	topics      []Topic
	lessons     []Lesson
	questions   []Question
}

// NewMemoryContentRepository creates a repository over copies of topics,
// lessons and questions.
func NewMemoryContentRepository(topics []Topic, lessons []Lesson, questions []Question) *MemoryContentRepository {
	return &MemoryContentRepository{
		topics:    cloneTopics(topics),
		lessons:   cloneLessons(lessons),
		questions: cloneQuestions(questions),
	}
}

func cloneTopics(topics []Topic) []Topic {
	return append([]Topic(nil), topics...)
}

func cloneLessons(lessons []Lesson) []Lesson {
	return append([]Lesson(nil), lessons...)
}

func cloneOptions(options []Option) []Option {
	return append([]Option(nil), options...)
}

func cloneQuestions(questions []Question) []Question {
	out := make([]Question, len(questions))
	for i, q := range questions {
		q.Options = cloneOptions(q.Options)
		out[i] = q
	}
	return out
}

func (r *MemoryContentRepository) ListPublishedTopics(_ context.Context) ([]Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, topic := range r.topics {
		if topic.Slug == slug && topic.Status == StatusPublished {
			return topic, nil
		}
	}
	return Topic{}, ErrNotFound
}

func (r *MemoryContentRepository) GetTopicBySlug(_ context.Context, slug string) (Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, topic := range r.topics {
		if topic.Slug == slug {
			return topic, nil
		}
	}
	return Topic{}, ErrNotFound
}

func (r *MemoryContentRepository) ListPublishedLessonsByTopicID(_ context.Context, topicID string) ([]Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, lesson := range r.lessons {
		if lesson.ID == id && lesson.Status == StatusPublished {
			return lesson, nil
		}
	}
	return Lesson{}, ErrNotFound
}

func (r *MemoryContentRepository) GetLessonByID(_ context.Context, id string) (Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, lesson := range r.lessons {
		if lesson.ID == id {
			return lesson, nil
		}
	}
	return Lesson{}, ErrNotFound
}

func (r *MemoryContentRepository) CreateDraftLesson(_ context.Context, topicID string, lesson Lesson) (Lesson, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextLesson++
	lesson.ID = fmt.Sprintf("lesson-%d", r.nextLesson)
	lesson.TopicID = topicID
	lesson.Status = StatusDraft
	r.lessons = append(r.lessons, lesson)
	return lesson, nil
}

func (r *MemoryContentRepository) UpdateDraftLesson(_ context.Context, lesson Lesson) (Lesson, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.lessons {
		if existing.ID == lesson.ID && existing.Status == StatusDraft {
			lesson.Status = StatusDraft
			r.lessons[i] = lesson
			return lesson, nil
		}
	}
	return Lesson{}, ErrNotFound
}

func (r *MemoryContentRepository) ListPublishedQuestionsByTopicID(_ context.Context, topicID string) ([]Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	questions := make([]Question, 0, len(r.questions))
	for _, question := range r.questions {
		if question.TopicID != topicID || question.Status != StatusPublished {
			continue
		}
		options := cloneOptions(question.Options)
		sort.SliceStable(options, func(i, j int) bool {
			return options[i].Position < options[j].Position
		})
		question.Options = options
		questions = append(questions, question)
	}
	return questions, nil
}

func (r *MemoryContentRepository) GetQuestionByID(_ context.Context, id string) (Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, question := range r.questions {
		if question.ID == id {
			q := question
			q.Options = cloneOptions(question.Options)
			return q, nil
		}
	}
	return Question{}, ErrNotFound
}

func (r *MemoryContentRepository) CreateDraftQuestion(_ context.Context, topicID string, question Question) (Question, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextQuest++
	question.ID = fmt.Sprintf("question-%d", r.nextQuest)
	question.TopicID = topicID
	question.Status = StatusDraft
	question.Options = r.assignOptionIDs(question.Options)
	r.questions = append(r.questions, question)
	out := question
	out.Options = cloneOptions(question.Options)
	return out, nil
}

func (r *MemoryContentRepository) UpdateDraftQuestion(_ context.Context, question Question) (Question, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.questions {
		if existing.ID == question.ID && existing.Status == StatusDraft {
			question.Status = StatusDraft
			question.Options = r.assignOptionIDs(question.Options)
			r.questions[i] = question
			out := question
			out.Options = cloneOptions(question.Options)
			return out, nil
		}
	}
	return Question{}, ErrNotFound
}

func (r *MemoryContentRepository) assignOptionIDs(options []Option) []Option {
	out := cloneOptions(options)
	for i := range out {
		r.nextOption++
		out[i].ID = fmt.Sprintf("option-%d", r.nextOption)
		if out[i].Position == 0 {
			out[i].Position = i + 1
		}
	}
	return out
}
