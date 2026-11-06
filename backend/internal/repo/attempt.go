package repo

import (
	"context"
	"errors"
	"time"
)

// ErrAttemptNotFound is returned when no stored attempt matches the lookup.
var ErrAttemptNotFound = errors.New("attempt not found")

// QuizAttemptAnswer is a snapshot of one answer in a stored attempt. The
// question and option texts are kept so later content edits do not change a
// past attempt.
type QuizAttemptAnswer struct {
	QuestionID         string
	QuestionText       string
	SelectedOptionText string
	CorrectOptionText  string
	IsCorrect          bool
}

// QuizAttempt is one stored quiz submission of a user.
type QuizAttempt struct {
	ID        string
	UserID    string
	TopicID   string
	TopicSlug string
	Score     int
	Total     int
	CreatedAt time.Time
	Answers   []QuizAttemptAnswer
}

// AttemptRepository stores quiz attempts per user. Every creation adds a new
// attempt; an existing attempt is never overwritten.
type AttemptRepository interface {
	CreateAttempt(ctx context.Context, attempt QuizAttempt) (QuizAttempt, error)
	ListAttemptsByUserID(ctx context.Context, userID string) ([]QuizAttempt, error)
	GetAttemptByID(ctx context.Context, id string) (QuizAttempt, error)
}
