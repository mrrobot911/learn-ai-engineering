// Package repo defines the content data access layer. SQL belongs only here.
package repo

import (
	"context"
	"errors"
)

// ContentStatus is the status of a topic or lesson. Only published content is
// visible to users.
type ContentStatus string

const (
	StatusDraft     ContentStatus = "draft"
	StatusPublished ContentStatus = "published"
)

// Topic is a topic of the platform.
type Topic struct {
	ID          string
	Slug        string
	Title       string
	Description string
	Position    int
	Status      ContentStatus
}

// Lesson is a lesson that belongs to a topic.
type Lesson struct {
	ID       string
	TopicID  string
	Title    string
	Body     string
	Position int
	Status   ContentStatus
}

// Option is one answer option of a question. Position defines the stable order
// the options are shown in; IsCorrect is never sent to a quiz client.
type Option struct {
	ID        string
	Text      string
	Position  int
	IsCorrect bool
}

// Question is a multiple-choice question that belongs to a topic.
type Question struct {
	ID          string
	TopicID     string
	Text        string
	Explanation string
	Options     []Option
	Status      ContentStatus
}

// ErrNotFound is returned when no published content matches the lookup.
var ErrNotFound = errors.New("content not found")

// ContentRepository reads published content. Implementations filter on status
// and return ErrNotFound when a requested topic or lesson does not exist or is
// not published.
type ContentRepository interface {
	ListPublishedTopics(ctx context.Context) ([]Topic, error)
	GetPublishedTopicBySlug(ctx context.Context, slug string) (Topic, error)
	ListPublishedLessonsByTopicID(ctx context.Context, topicID string) ([]Lesson, error)
	GetPublishedLessonByID(ctx context.Context, id string) (Lesson, error)
	ListPublishedQuestionsByTopicID(ctx context.Context, topicID string) ([]Question, error)
}
