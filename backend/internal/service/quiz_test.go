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
	}, repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(42)))
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
	), repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(7)))

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
	}}, repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(1)))

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
	}}, repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(1)))

	_, err := s.StartQuiz(context.Background(), "topic")
	if !errors.Is(err, cause) {
		t.Errorf("expected the repository error to be wrapped, got %v", err)
	}
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		t.Errorf("expected an unexpected repository error to stay unmapped, got %v", err)
	}
}

func gradingQuestions() []repo.Question {
	questions := []repo.Question{
		{ID: "q1", TopicID: "topic-1", Explanation: "Because 1.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q1-correct", Text: "Right 1", Position: 1, IsCorrect: true},
			{ID: "q1-wrong", Text: "Wrong 1", Position: 2},
		}},
		{ID: "q2", TopicID: "topic-1", Explanation: "Because 2.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q2-correct", Text: "Right 2", Position: 1, IsCorrect: true},
			{ID: "q2-wrong", Text: "Wrong 2", Position: 2},
		}},
		{ID: "q3", TopicID: "topic-1", Explanation: "Because 3.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q3-correct", Text: "Right 3", Position: 1, IsCorrect: true},
			{ID: "q3-wrong", Text: "Wrong 3", Position: 2},
		}},
		{ID: "q-draft", TopicID: "topic-1", Explanation: "Hidden.", Status: repo.StatusDraft, Options: []repo.Option{
			{ID: "q-draft-correct", Text: "Draft right", Position: 1, IsCorrect: true},
		}},
	}
	return questions
}

func TestSubmitQuizGradesAnswers(t *testing.T) {
	questions := gradingQuestions()
	s := seededService(questions)
	correct := "q1-correct"
	foreignOption := "q2-correct" // belongs to q2, not q3

	result, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{
		{QuestionID: "q1", OptionID: &correct},
		{QuestionID: "q2", OptionID: nil},
		{QuestionID: "q3", OptionID: &foreignOption},
	}, "")
	if err != nil {
		t.Fatalf("submit quiz: %v", err)
	}
	if result.Score != 1 || result.Total != 3 {
		t.Fatalf("expected score 1 of 3, got %d of %d", result.Score, result.Total)
	}
	if len(result.Answers) != 3 {
		t.Fatalf("expected 3 graded answers, got %d", len(result.Answers))
	}

	for i, want := range []QuizAnswerResult{
		{QuestionID: "q1", IsCorrect: true, CorrectOptionText: "Right 1", Explanation: "Because 1."},
		{QuestionID: "q2", IsCorrect: false, CorrectOptionText: "Right 2", Explanation: "Because 2."},
		{QuestionID: "q3", IsCorrect: false, CorrectOptionText: "Right 3", Explanation: "Because 3."},
	} {
		if result.Answers[i] != want {
			t.Errorf("answer %d: expected %+v, got %+v", i, want, result.Answers[i])
		}
	}
}

func TestSubmitQuizRejectsQuestionNotBelongingToTopic(t *testing.T) {
	s := seededService(gradingQuestions())
	foreign := publishedQuestions("topic-2", 1)[0]

	for _, tt := range []struct {
		name       string
		questionID string
	}{
		{"question of another topic", foreign.ID},
		{"unknown question", "unknown-question"},
		{"draft question of the topic", "q-draft"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{
				{QuestionID: tt.questionID, OptionID: nil},
				{QuestionID: "q1", OptionID: nil},
			}, "")
			if !hasCode(err, apperr.CodeValidation) {
				t.Errorf("expected validation_error, got %v", err)
			}
		})
	}
}

func TestSubmitQuizIgnoresDuplicateAnswers(t *testing.T) {
	for _, tt := range []struct {
		name        string
		first       string
		second      string
		wantCorrect bool
	}{
		{"first occurrence counts over the correct duplicate", "q1-wrong", "q1-correct", false},
		{"first correct duplicate is kept", "q1-correct", "q1-wrong", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := seededService(gradingQuestions())
			first, second := tt.first, tt.second

			result, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{
				{QuestionID: "q1", OptionID: &first},
				{QuestionID: "q1", OptionID: &second},
			}, "")
			if err != nil {
				t.Fatalf("submit quiz: %v", err)
			}
			if result.Total != 1 || len(result.Answers) != 1 {
				t.Fatalf("expected one graded answer, got %d of %d", result.Score, result.Total)
			}
			if result.Answers[0].IsCorrect != tt.wantCorrect {
				t.Errorf("expected is_correct %v, got %v", tt.wantCorrect, result.Answers[0].IsCorrect)
			}
		})
	}
}

func TestSubmitQuizRejectsEmptyAnswerList(t *testing.T) {
	s := seededService(gradingQuestions())

	for name, answers := range map[string][]QuizAnswer{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			_, err := s.SubmitQuiz(context.Background(), "topic", answers, "")
			if !hasCode(err, apperr.CodeValidation) {
				t.Errorf("expected validation_error, got %v", err)
			}
		})
	}
}

func TestSubmitQuizTopicNotFound(t *testing.T) {
	s := NewContentServiceWithRandom(unfilteredRepository{topics: []repo.Topic{
		{ID: "topic-1", Slug: "published", Status: repo.StatusPublished},
		{ID: "topic-2", Slug: "draft", Status: repo.StatusDraft},
	}}, repo.NewMemoryAttemptRepository(), rand.New(rand.NewSource(1)))

	for _, slug := range []string{"unknown", "draft"} {
		t.Run(slug, func(t *testing.T) {
			_, err := s.SubmitQuiz(context.Background(), slug, []QuizAnswer{{QuestionID: "q1"}}, "")
			if !hasCode(err, apperr.CodeNotFound) {
				t.Errorf("expected not_found, got %v", err)
			}
		})
	}
}

type gradingQuestionsFailingRepository struct {
	unfilteredRepository
	err error
}

func (r gradingQuestionsFailingRepository) ListPublishedQuestionsByTopicID(context.Context, string) ([]repo.Question, error) {
	return nil, r.err
}

func TestSubmitQuizWrapsRepositoryError(t *testing.T) {
	cause := errors.New("query failed")
	s := NewContentService(gradingQuestionsFailingRepository{
		unfilteredRepository: unfilteredRepository{topics: []repo.Topic{{ID: "topic-1", Slug: "topic", Status: repo.StatusPublished}}},
		err:                  cause,
	}, repo.NewMemoryAttemptRepository())

	_, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{{QuestionID: "q1"}}, "")
	if !errors.Is(err, cause) {
		t.Errorf("expected the repository error to be wrapped, got %v", err)
	}
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		t.Errorf("expected an unexpected repository error to stay unmapped, got %v", err)
	}
}
