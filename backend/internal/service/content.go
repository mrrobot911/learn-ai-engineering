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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// quizSize is the number of questions a quiz returns when the topic has enough
// published questions. Version 1 does not let the client choose it.
const quizSize = 5

func quizTracer() trace.Tracer {
	return otel.Tracer("github.com/mrrobot911/learn-ai-engineering/internal/service")
}

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

	loadCtx, loadSpan := quizTracer().Start(ctx, "quiz.load_questions")
	questions, err := s.repository.ListPublishedQuestionsByTopicID(loadCtx, topic.ID)
	if err != nil {
		loadSpan.End()
		return QuizResult{}, fmt.Errorf("list questions of topic %q for grading: %w", slug, err)
	}
	loadSpan.End()

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

	_, gradeSpan := quizTracer().Start(ctx, "quiz.grade")
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
	gradeSpan.End()

	result := QuizResult{
		Score:   score,
		Total:   len(results),
		Answers: results,
	}
	if userID != "" {
		storeCtx, storeSpan := quizTracer().Start(ctx, "quiz.store_attempt")
		if _, err := s.attempts.CreateAttempt(storeCtx, repo.QuizAttempt{
			UserID:    userID,
			TopicID:   topic.ID,
			TopicSlug: topic.Slug,
			Score:     result.Score,
			Total:     result.Total,
			CreatedAt: time.Now(),
			Answers:   attemptAnswers,
		}); err != nil {
			storeSpan.End()
			return QuizResult{}, fmt.Errorf("store quiz attempt for user %q: %w", userID, err)
		}
		storeSpan.End()
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

// DraftLessonInput is the data for creating or updating a draft lesson.
type DraftLessonInput struct {
	Title    string
	Body     string
	Position int
}

// DraftQuestionOptionInput is one option of a draft question.
type DraftQuestionOptionInput struct {
	Text      string
	IsCorrect bool
}

// DraftQuestionInput is the data for creating or updating a draft question.
type DraftQuestionInput struct {
	Text    string
	Options []DraftQuestionOptionInput
}

// CreateDraftLesson creates a draft lesson under the topic.
func (s *ContentService) CreateDraftLesson(ctx context.Context, topicSlug string, input DraftLessonInput) (repo.Lesson, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	lesson, err := s.repository.CreateDraftLesson(ctx, topic.ID, repo.Lesson{
		Title:    input.Title,
		Body:     input.Body,
		Position: input.Position,
	})
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("create draft lesson for topic %q: %w", topicSlug, err)
	}
	return lesson, nil
}

// UpdateDraftLesson updates an existing draft lesson under the topic.
func (s *ContentService) UpdateDraftLesson(ctx context.Context, topicSlug, lessonID string, input DraftLessonInput) (repo.Lesson, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	existing, err := s.repository.GetLessonByID(ctx, lessonID)
	if errors.Is(err, repo.ErrNotFound) || existing.TopicID != topic.ID || existing.Status != repo.StatusDraft {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get lesson %q: %w", lessonID, err)
	}

	updated, err := s.repository.UpdateDraftLesson(ctx, repo.Lesson{
		ID:       lessonID,
		TopicID:  topic.ID,
		Title:    input.Title,
		Body:     input.Body,
		Position: input.Position,
		Status:   repo.StatusDraft,
	})
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("update draft lesson %q: %w", lessonID, err)
	}
	return updated, nil
}

// CreateDraftQuestion creates a draft question under the topic.
func (s *ContentService) CreateDraftQuestion(ctx context.Context, topicSlug string, input DraftQuestionInput) (repo.Question, error) {
	if err := validateDraftQuestion(input); err != nil {
		return repo.Question{}, err
	}

	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	question, err := s.repository.CreateDraftQuestion(ctx, topic.ID, draftQuestionFromInput(input))
	if err != nil {
		return repo.Question{}, fmt.Errorf("create draft question for topic %q: %w", topicSlug, err)
	}
	return question, nil
}

// UpdateDraftQuestion updates an existing draft question under the topic.
func (s *ContentService) UpdateDraftQuestion(ctx context.Context, topicSlug, questionID string, input DraftQuestionInput) (repo.Question, error) {
	if err := validateDraftQuestion(input); err != nil {
		return repo.Question{}, err
	}

	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	existing, err := s.repository.GetQuestionByID(ctx, questionID)
	if errors.Is(err, repo.ErrNotFound) || existing.TopicID != topic.ID || existing.Status != repo.StatusDraft {
		return repo.Question{}, notFound("Question not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("get question %q: %w", questionID, err)
	}

	q := draftQuestionFromInput(input)
	q.ID = questionID
	q.TopicID = topic.ID
	q.Status = repo.StatusDraft

	updated, err := s.repository.UpdateDraftQuestion(ctx, q)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Question not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("update draft question %q: %w", questionID, err)
	}
	return updated, nil
}

func draftQuestionFromInput(input DraftQuestionInput) repo.Question {
	options := make([]repo.Option, 0, len(input.Options))
	for i, opt := range input.Options {
		options = append(options, repo.Option{
			Text:      opt.Text,
			Position:  i + 1,
			IsCorrect: opt.IsCorrect,
		})
	}
	return repo.Question{
		Text:    input.Text,
		Options: options,
	}
}

func validateDraftQuestion(input DraftQuestionInput) error {
	if len(input.Options) < 2 || len(input.Options) > 6 {
		return validationError("A question must have between 2 and 6 options.")
	}
	correct := 0
	for _, opt := range input.Options {
		if opt.IsCorrect {
			correct++
		}
	}
	if correct != 1 {
		return validationError("A question must have exactly one correct option.")
	}
	return nil
}

func notFound(message string) error {
	return &apperr.Error{Code: apperr.CodeNotFound, Message: message}
}

// PublishLesson publishes a lesson, making it visible through public endpoints.
func (s *ContentService) PublishLesson(ctx context.Context, topicSlug, lessonID string) (repo.Lesson, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	lesson, err := s.repository.PublishLesson(ctx, topic.ID, lessonID)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("publish lesson %q: %w", lessonID, err)
	}
	return lesson, nil
}

// UnpublishLesson unpublishes a lesson, returning it to draft status.
func (s *ContentService) UnpublishLesson(ctx context.Context, topicSlug, lessonID string) (repo.Lesson, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	lesson, err := s.repository.UnpublishLesson(ctx, topic.ID, lessonID)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Lesson{}, notFound("Lesson not found.")
	}
	if err != nil {
		return repo.Lesson{}, fmt.Errorf("unpublish lesson %q: %w", lessonID, err)
	}
	return lesson, nil
}

// PublishQuestion publishes a question, making it visible through public endpoints.
func (s *ContentService) PublishQuestion(ctx context.Context, topicSlug, questionID string) (repo.Question, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	question, err := s.repository.PublishQuestion(ctx, topic.ID, questionID)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Question not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("publish question %q: %w", questionID, err)
	}
	return question, nil
}

// UnpublishQuestion unpublishes a question, returning it to draft status.
func (s *ContentService) UnpublishQuestion(ctx context.Context, topicSlug, questionID string) (repo.Question, error) {
	topic, err := s.repository.GetTopicBySlug(ctx, topicSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Topic not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("get topic %q: %w", topicSlug, err)
	}

	question, err := s.repository.UnpublishQuestion(ctx, topic.ID, questionID)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Question{}, notFound("Question not found.")
	}
	if err != nil {
		return repo.Question{}, fmt.Errorf("unpublish question %q: %w", questionID, err)
	}
	return question, nil
}
