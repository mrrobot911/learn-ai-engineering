package service

import (
	"context"
	"testing"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
)

func attemptQuestions(suffix string) []repo.Question {
	return []repo.Question{
		{ID: "q1", TopicID: "topic-1", Text: "Question 1" + suffix, Explanation: "Because 1.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q1-correct", Text: "Right 1" + suffix, Position: 1, IsCorrect: true},
			{ID: "q1-wrong", Text: "Wrong 1" + suffix, Position: 2},
		}},
		{ID: "q2", TopicID: "topic-1", Text: "Question 2" + suffix, Explanation: "Because 2.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q2-correct", Text: "Right 2" + suffix, Position: 1, IsCorrect: true},
			{ID: "q2-wrong", Text: "Wrong 2" + suffix, Position: 2},
		}},
		{ID: "q3", TopicID: "topic-1", Text: "Question 3" + suffix, Explanation: "Because 3.", Status: repo.StatusPublished, Options: []repo.Option{
			{ID: "q3-correct", Text: "Right 3" + suffix, Position: 1, IsCorrect: true},
			{ID: "q3-wrong", Text: "Wrong 3" + suffix, Position: 2},
		}},
	}
}

func attemptService(questions []repo.Question, attempts repo.AttemptRepository) *ContentService {
	return NewContentService(unfilteredRepository{
		topics:    []repo.Topic{{ID: "topic-1", Slug: "topic", Status: repo.StatusPublished}},
		questions: questions,
	}, attempts)
}

func attemptAnswers() []QuizAnswer {
	correct := "q1-correct"
	foreign := "q2-correct" // belongs to q2, not q3
	return []QuizAnswer{
		{QuestionID: "q1", OptionID: &correct},
		{QuestionID: "q2", OptionID: nil},
		{QuestionID: "q3", OptionID: &foreign},
	}
}

func TestSubmitQuizStoresAttemptForLoggedInUser(t *testing.T) {
	attempts := repo.NewMemoryAttemptRepository()
	s := attemptService(attemptQuestions(""), attempts)

	before := time.Now()
	result, err := s.SubmitQuiz(context.Background(), "topic", attemptAnswers(), "user-1")
	if err != nil {
		t.Fatalf("submit quiz: %v", err)
	}
	after := time.Now()

	if result.Score != 1 || result.Total != 3 {
		t.Fatalf("expected score 1 of 3, got %d of %d", result.Score, result.Total)
	}

	stored, err := s.ListAttempts(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list attempts: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected one stored attempt, got %d: %v", len(stored), stored)
	}
	attempt := stored[0]
	if attempt.ID == "" {
		t.Error("expected the stored attempt to have an id")
	}
	if attempt.UserID != "user-1" || attempt.TopicID != "topic-1" || attempt.TopicSlug != "topic" {
		t.Errorf("unexpected attempt owner or topic: %+v", attempt)
	}
	if attempt.Score != 1 || attempt.Total != 3 {
		t.Errorf("expected score 1 of 3 in the attempt, got %d of %d", attempt.Score, attempt.Total)
	}
	if attempt.CreatedAt.Before(before) || attempt.CreatedAt.After(after) {
		t.Errorf("expected the attempt time between %s and %s, got %s", before, after, attempt.CreatedAt)
	}
	if len(attempt.Answers) != 3 {
		t.Fatalf("expected one answer row per question, got %d", len(attempt.Answers))
	}
}

func TestSubmitQuizCreatesANewAttemptPerSubmission(t *testing.T) {
	attempts := repo.NewMemoryAttemptRepository()
	s := attemptService(attemptQuestions(""), attempts)

	correct := "q1-correct"
	if _, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{{QuestionID: "q1", OptionID: &correct}}, "user-1"); err != nil {
		t.Fatalf("first submission: %v", err)
	}
	if _, err := s.SubmitQuiz(context.Background(), "topic", []QuizAnswer{{QuestionID: "q2", OptionID: nil}}, "user-1"); err != nil {
		t.Fatalf("second submission: %v", err)
	}

	stored, err := s.ListAttempts(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list attempts: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected two stored attempts, got %d: %v", len(stored), stored)
	}
	if stored[0].ID == stored[1].ID {
		t.Errorf("expected distinct attempt ids, got %q twice", stored[0].ID)
	}

	questionIDs := make(map[string]bool, len(stored))
	for _, attempt := range stored {
		if attempt.Total != 1 || len(attempt.Answers) != 1 {
			t.Errorf("expected each attempt to keep its own answer, got %+v", attempt)
		}
		questionIDs[attempt.Answers[0].QuestionID] = true
	}
	if !questionIDs["q1"] || !questionIDs["q2"] {
		t.Errorf("expected one attempt for q1 and one for q2, got %v", questionIDs)
	}
}

func TestSubmitQuizKeepsSnapshotWording(t *testing.T) {
	attempts := repo.NewMemoryAttemptRepository()

	firstService := attemptService(attemptQuestions(" v1"), attempts)
	if _, err := firstService.SubmitQuiz(context.Background(), "topic", attemptAnswers(), "user-1"); err != nil {
		t.Fatalf("first submission: %v", err)
	}
	first, err := firstService.ListAttempts(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list first attempt: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected one first attempt, got %d", len(first))
	}

	secondService := attemptService(attemptQuestions(" v2"), attempts)
	if _, err := secondService.SubmitQuiz(context.Background(), "topic", attemptAnswers(), "user-1"); err != nil {
		t.Fatalf("second submission: %v", err)
	}

	kept, err := secondService.GetAttempt(context.Background(), "user-1", first[0].ID)
	if err != nil {
		t.Fatalf("get first attempt: %v", err)
	}
	want := []repo.QuizAttemptAnswer{
		{QuestionID: "q1", QuestionText: "Question 1 v1", SelectedOptionText: "Right 1 v1", CorrectOptionText: "Right 1 v1", IsCorrect: true},
		{QuestionID: "q2", QuestionText: "Question 2 v1", SelectedOptionText: "", CorrectOptionText: "Right 2 v1", IsCorrect: false},
		{QuestionID: "q3", QuestionText: "Question 3 v1", SelectedOptionText: "", CorrectOptionText: "Right 3 v1", IsCorrect: false},
	}
	if len(kept.Answers) != len(want) {
		t.Fatalf("expected %d snapshot answers, got %d", len(want), len(kept.Answers))
	}
	for i, wantAnswer := range want {
		if kept.Answers[i] != wantAnswer {
			t.Errorf("answer %d: expected snapshot %+v, got %+v", i, wantAnswer, kept.Answers[i])
		}
	}

	all, err := secondService.ListAttempts(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list attempts after the second submission: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected two attempts, got %d", len(all))
	}
	var newAttempt repo.QuizAttempt
	for _, attempt := range all {
		if attempt.ID != first[0].ID {
			newAttempt = attempt
		}
	}
	if newAttempt.ID == "" {
		t.Fatal("expected a new attempt next to the first one")
	}
	if newAttempt.Answers[0].QuestionText != "Question 1 v2" || newAttempt.Answers[0].CorrectOptionText != "Right 1 v2" {
		t.Errorf("expected the new attempt to use the new wording, got %+v", newAttempt.Answers[0])
	}
}

func TestSubmitQuizWithoutSessionStoresNothing(t *testing.T) {
	attempts := repo.NewMemoryAttemptRepository()
	s := attemptService(attemptQuestions(""), attempts)

	result, err := s.SubmitQuiz(context.Background(), "topic", attemptAnswers(), "")
	if err != nil {
		t.Fatalf("anonymous submission: %v", err)
	}
	if result.Score != 1 || result.Total != 3 {
		t.Fatalf("expected the anonymous submission to be graded, got %d of %d", result.Score, result.Total)
	}

	stored, err := attempts.ListAttemptsByUserID(context.Background(), "")
	if err != nil {
		t.Fatalf("list anonymous attempts: %v", err)
	}
	if len(stored) != 0 {
		t.Errorf("expected nothing to be stored for an anonymous submission, got %v", stored)
	}
}

func TestListAttemptsReturnsOnlyTheUsersAttemptsNewestFirst(t *testing.T) {
	attempts := repo.NewMemoryAttemptRepository()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	older, err := attempts.CreateAttempt(context.Background(), repo.QuizAttempt{
		UserID: "user-1", TopicID: "topic-1", TopicSlug: "topic", Score: 1, Total: 2, CreatedAt: base,
	})
	if err != nil {
		t.Fatalf("store older attempt: %v", err)
	}
	newer, err := attempts.CreateAttempt(context.Background(), repo.QuizAttempt{
		UserID: "user-1", TopicID: "topic-1", TopicSlug: "topic", Score: 2, Total: 2, CreatedAt: base.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("store newer attempt: %v", err)
	}
	if _, err := attempts.CreateAttempt(context.Background(), repo.QuizAttempt{
		UserID: "user-2", TopicID: "topic-1", TopicSlug: "topic", Score: 0, Total: 2, CreatedAt: base.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("store other user's attempt: %v", err)
	}

	s := attemptService(attemptQuestions(""), attempts)
	stored, err := s.ListAttempts(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list attempts: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected only user-1's two attempts, got %d: %v", len(stored), stored)
	}
	if stored[0].ID != newer.ID || stored[1].ID != older.ID {
		t.Errorf("expected newest first (%q, %q), got (%q, %q)", newer.ID, older.ID, stored[0].ID, stored[1].ID)
	}
	for _, attempt := range stored {
		if attempt.UserID != "user-1" {
			t.Errorf("expected only user-1's attempts, got %+v", attempt)
		}
	}

	if _, err := s.GetAttempt(context.Background(), "user-2", older.ID); !hasCode(err, apperr.CodeNotFound) {
		t.Errorf("expected not_found for another user's attempt, got %v", err)
	}
	if _, err := s.GetAttempt(context.Background(), "user-1", "missing"); !hasCode(err, apperr.CodeNotFound) {
		t.Errorf("expected not_found for a missing attempt, got %v", err)
	}
}
