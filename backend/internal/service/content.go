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

// QuizAnswer is one answer of a quiz submission. OptionID is nil when the
// question is unanswered.
type QuizAnswer struct {
	QuestionID string
	OptionID   *string
}

// QuizAnswerResult is the graded answer of one question of the quiz.
type QuizAnswerResult struct {
	QuestionID        string
	IsCorrect         bool
	CorrectOptionText string
	Explanation       string
}

// QuizResult is the graded result of a quiz submission.
type QuizResult struct {
	Score   int
	Total   int
	Answers []QuizAnswerResult
}

// ContentService serves published topics and lessons, starts and grades
// quizzes and stores the attempts of logged-in users.
type ContentService struct {
	repository repo.ContentRepository
	attempts   repo.AttemptRepository
	random     *rand.Rand
	randomMu   sync.Mutex
}

// NewContentService creates a content service over repository and attempts.
func NewContentService(repository repo.ContentRepository, attempts repo.AttemptRepository) *ContentService {
	return NewContentServiceWithRandom(repository, attempts, rand.New(rand.NewSource(time.Now().UnixNano())))
}

// NewContentServiceWithRandom creates a content service that draws its quiz
// questions from random, so tests can pin the sequence.
func NewContentServiceWithRandom(repository repo.ContentRepository, attempts repo.AttemptRepository, random *rand.Rand) *ContentService {
	return &ContentService{repository: repository, attempts: attempts, random: random}
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

// SubmitQuiz grades the submission for a topic. A submission containing a
// question that does not belong to the topic is rejected with validation_error
// and nothing is graded. An unanswered question or an option that does not
// belong to the question counts as incorrect. The first occurrence of a
// question id counts; later duplicates are ignored. An empty answer list is a
// validation_error. A non-empty userID stores the attempt for that user; an
// anonymous submission stores nothing.
func (s *ContentService) SubmitQuiz(ctx context.Context, slug string, answers []QuizAnswer, userID string) (QuizResult, error) {
	if len(answers) == 0 {
		return QuizResult{}, validationError("The submission must include at least one answer.")
	}

	topic, err := s.getPublishedTopic(ctx, slug)
	if err != nil {
		return QuizResult{}, err
	}

	questions, err := s.repository.ListPublishedQuestionsByTopicID(ctx, topic.ID)
	if err != nil {
		return QuizResult{}, fmt.Errorf("list questions of topic %q for grading: %w", slug, err)
	}

	published := make(map[string]repo.Question, len(questions))
	for _, question := range questions {
		if question.TopicID == topic.ID && question.Status == repo.StatusPublished {
			published[question.ID] = question
		}
	}

	submitted := make([]gradedQuestion, 0, len(answers))
	seen := make(map[string]bool, len(answers))
	for _, answer := range answers {
		if seen[answer.QuestionID] {
			continue
		}
		question, ok := published[answer.QuestionID]
		if !ok {
			return QuizResult{}, validationError("A submission cannot include a question that does not belong to the topic.")
		}
		submitted = append(submitted, gradedQuestion{question: question, answer: answer})
		seen[answer.QuestionID] = true
	}

	results := make([]QuizAnswerResult, 0, len(submitted))
	attemptAnswers := make([]repo.QuizAttemptAnswer, 0, len(submitted))
	score := 0
	for _, graded := range submitted {
		question := graded.question
		correctOption := repo.Option{}
		for _, option := range question.Options {
			if option.IsCorrect {
				correctOption = option
				break
			}
		}

		selectedOptionText := ""
		isCorrect := false
		if graded.answer.OptionID != nil {
			for _, option := range question.Options {
				if option.ID == *graded.answer.OptionID {
					selectedOptionText = option.Text
					isCorrect = option.IsCorrect
					break
				}
			}
		}

		if isCorrect {
			score++
		}
		results = append(results, QuizAnswerResult{
			QuestionID:        question.ID,
			IsCorrect:         isCorrect,
			CorrectOptionText: correctOption.Text,
			Explanation:       question.Explanation,
		})
		attemptAnswers = append(attemptAnswers, repo.QuizAttemptAnswer{
			QuestionID:         question.ID,
			QuestionText:       question.Text,
			SelectedOptionText: selectedOptionText,
			CorrectOptionText:  correctOption.Text,
			IsCorrect:          isCorrect,
		})
	}

	result := QuizResult{
		Score:   score,
		Total:   len(results),
		Answers: results,
	}
	if userID != "" {
		if _, err := s.attempts.CreateAttempt(ctx, repo.QuizAttempt{
			UserID:    userID,
			TopicID:   topic.ID,
			TopicSlug: topic.Slug,
			Score:     result.Score,
			Total:     result.Total,
			CreatedAt: time.Now(),
			Answers:   attemptAnswers,
		}); err != nil {
			return QuizResult{}, fmt.Errorf("store quiz attempt for user %q: %w", userID, err)
		}
	}
	return result, nil
}

// ListAttempts returns the stored quiz attempts of the user, newest first. The
// result is never nil.
func (s *ContentService) ListAttempts(ctx context.Context, userID string) ([]repo.QuizAttempt, error) {
	attempts, err := s.attempts.ListAttemptsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list quiz attempts of user %q: %w", userID, err)
	}
	if attempts == nil {
		attempts = []repo.QuizAttempt{}
	}
	return attempts, nil
}

// GetAttempt returns one stored attempt of the user. An attempt of another user
// is reported as not found.
func (s *ContentService) GetAttempt(ctx context.Context, userID, attemptID string) (repo.QuizAttempt, error) {
	attempt, err := s.attempts.GetAttemptByID(ctx, attemptID)
	if errors.Is(err, repo.ErrAttemptNotFound) {
		return repo.QuizAttempt{}, notFound("Attempt not found.")
	}
	if err != nil {
		return repo.QuizAttempt{}, fmt.Errorf("get quiz attempt %q: %w", attemptID, err)
	}
	if attempt.UserID != userID {
		return repo.QuizAttempt{}, notFound("Attempt not found.")
	}
	return attempt, nil
}

type gradedQuestion struct {
	question repo.Question
	answer   QuizAnswer
}

func validationError(message string) error {
	return &apperr.Error{Code: apperr.CodeValidation, Message: message}
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
