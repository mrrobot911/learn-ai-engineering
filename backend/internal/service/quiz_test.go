package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

func publishedQuestions(topicID string, count int) []repo.Question {
	questions := make([]repo.Question, 0, count)
	for i := 1; i <= count; i++ {
		questions = append(questions, repo.Question{
			ID:      fmt.Sprintf("%s-question-%d", topicID, i),
			TopicID: topicID,
			Text:    fmt.Sprintf("Question %d?", i),
			Status:  repo.StatusPublished,
			Options: []repo.Option{
				{ID: fmt.Sprintf("%s-question-%d-option-1", topicID, i), Text: "First", Position: 1, IsCorrect: true},
				{ID: fmt.Sprintf("%s-question-%d-option-2", topicID, i), Text: "Second", Position: 2},
			},
		})
	}
	return questions
}

func seededService(questions []repo.Question) *ContentService {
	return NewContentServiceWithRandom(unfilteredRepository{
		topics:    []repo.Topic{{ID: "topic-1", Slug: "topic", Status: repo.StatusPublished}},
		questions: questions,
	}, rand.New(rand.NewSource(42)))
}

func TestStartQuizReturnsFiveQuestions(t *testing.T) {
	s := seededService(publishedQuestions("topic-1", 9))

	questions, err := s.StartQuiz(context.Background(), "topic")
	if err != nil {
		t.Fatalf("start quiz: %v", err)
	}
	if len(questions) != 5 {
		t.Fatalf("expected 5 questions, got %d: %v", len(questions), questions)
	}

	seen := make(map[string]bool, len(questions))
	for _, question := range questions {
		if question.TopicID != "topic-1" {
			t.Errorf("expected a question of topic-1, got %q", question.TopicID)
		}
		if seen[question.ID] {
			t.Errorf("question %q returned twice, so selection is not without replacement", question.ID)
		}
		seen[question.ID] = true
	}
}

func TestStartQuizReturnsAllQuestionsWhenFewerThanFive(t *testing.T) {
	s := seededService(publishedQuestions("topic-1", 3))

	questions, err := s.StartQuiz(context.Background(), "topic")
	if err != nil {
		t.Fatalf("start quiz: %v", err)
	}
	if len(questions) != 3 {
		t.Fatalf("expected all 3 published questions, got %d: %v", len(questions), questions)
	}
}

func TestStartQuizReturnsEmptySliceWithoutQuestions(t *testing.T) {
	s := seededService(nil)

	questions, err := s.StartQuiz(context.Background(), "topic")
	if err != nil {
		t.Fatalf("start quiz: %v", err)
	}
	if questions == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(questions) != 0 {
		t.Errorf("expected no questions, got %v", questions)
	}
}

func TestStartQuizExcludesDraftAndForeignTopicQuestions(t *testing.T) {
	questions := append(publishedQuestions("topic-1", 6),
		repo.Question{ID: "draft-question", TopicID: "topic-1", Status: repo.StatusDraft},
		repo.Question{ID: "foreign-question", TopicID: "topic-2", Status: repo.StatusPublished},
	)
	s := seededService(questions)

	for i := 0; i < 20; i++ {
		quiz, err := s.StartQuiz(context.Background(), "topic")
		if err != nil {
			t.Fatalf("start quiz: %v", err)
		}
		if len(quiz) != 5 {
			t.Fatalf("expected 5 published questions, got %d: %v", len(quiz), quiz)
		}
		for _, question := range quiz {
			if question.ID == "draft-question" {
				t.Fatal("a draft question was returned in a quiz")
			}
			if question.ID == "foreign-question" || question.TopicID != "topic-1" {
				t.Fatalf("a question of another topic was returned: %v", question)
			}
		}
	}
}

func TestStartQuizCanReturnDifferentQuestionSets(t *testing.T) {
	s := seededService(publishedQuestions("topic-1", 6))

	sets := make(map[string]bool)
	for i := 0; i < 20; i++ {
		quiz, err := s.StartQuiz(context.Background(), "topic")
		if err != nil {
			t.Fatalf("start quiz: %v", err)
		}
		if len(quiz) != 5 {
			t.Fatalf("expected 5 questions, got %d: %v", len(quiz), quiz)
		}
		ids := make([]string, 0, len(quiz))
		for _, question := range quiz {
			ids = append(ids, question.ID)
		}
		sets[strings.Join(ids, ",")] = true
	}
	if len(sets) < 2 {
		t.Errorf("expected at least two distinct question sets in 20 quizzes with a fixed seed, got %d", len(sets))
	}
}

func TestStartQuizKeepsOptionPositionOrder(t *testing.T) {
	s := NewContentServiceWithRandom(repo.NewMemoryContentRepository(
		[]repo.Topic{{ID: "topic-1", Slug: "topic", Status: repo.StatusPublished}},
		nil,
		[]repo.Question{{
			ID:      "question-1",
			TopicID: "topic-1",
			Status:  repo.StatusPublished,
			Options: []repo.Option{
				{ID: "option-3", Text: "Third", Position: 3},
				{ID: "option-1", Text: "First", Position: 1, IsCorrect: true},
				{ID: "option-4", Text: "Fourth", Position: 4},
				{ID: "option-2", Text: "Second", Position: 2},
			},
		}},
	), rand.New(rand.NewSource(7)))

	want := []string{"option-1", "option-2", "option-3", "option-4"}
	for call := 0; call < 2; call++ {
		quiz, err := s.StartQuiz(context.Background(), "topic")
		if err != nil {
			t.Fatalf("start quiz: %v", err)
		}
		if len(quiz) != 1 {
			t.Fatalf("expected the single question, got %d: %v", len(quiz), quiz)
		}
		if len(quiz[0].Options) != len(want) {
			t.Fatalf("expected %d options, got %d: %v", len(want), len(quiz[0].Options), quiz[0].Options)
		}
		for i, id := range want {
			if quiz[0].Options[i].ID != id {
				t.Errorf("call %d: expected option %q at position %d, got %q", call, id, i, quiz[0].Options[i].ID)
			}
		}
	}
}

func TestStartQuizNotFound(t *testing.T) {
	s := NewContentServiceWithRandom(unfilteredRepository{topics: []repo.Topic{
		{ID: "topic-1", Slug: "published", Status: repo.StatusPublished},
		{ID: "topic-2", Slug: "draft", Status: repo.StatusDraft},
	}}, rand.New(rand.NewSource(1)))

	for _, slug := range []string{"unknown", "draft"} {
		t.Run(slug, func(t *testing.T) {
			if _, err := s.StartQuiz(context.Background(), slug); !hasCode(err, apperr.CodeNotFound) {
				t.Errorf("expected not_found, got %v", err)
			}
		})
	}
}

// questionFailingRepository lists topics fine and fails only when the
// questions are read.
type questionFailingRepository struct {
	unfilteredRepository
}

func (r questionFailingRepository) ListPublishedQuestionsByTopicID(context.Context, string) ([]repo.Question, error) {
	return nil, r.err
}

func TestStartQuizWrapsRepositoryError(t *testing.T) {
	cause := errors.New("query failed")
	s := NewContentServiceWithRandom(questionFailingRepository{unfilteredRepository{
		topics: []repo.Topic{{ID: "topic-1", Slug: "topic", Status: repo.StatusPublished}},
		err:    cause,
	}}, rand.New(rand.NewSource(1)))

	_, err := s.StartQuiz(context.Background(), "topic")
	if !errors.Is(err, cause) {
		t.Errorf("expected the repository error to be wrapped, got %v", err)
	}
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		t.Errorf("expected an unexpected repository error to stay unmapped, got %v", err)
	}
}
