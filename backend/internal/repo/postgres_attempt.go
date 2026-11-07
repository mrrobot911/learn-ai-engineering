package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAttemptRepository is a PostgreSQL implementation of AttemptRepository.
type PostgresAttemptRepository struct {
	pool *pgxpool.Pool
}

var _ AttemptRepository = (*PostgresAttemptRepository)(nil)

// NewPostgresAttemptRepository creates an attempt repository backed by pool.
func NewPostgresAttemptRepository(pool *pgxpool.Pool) *PostgresAttemptRepository {
	return &PostgresAttemptRepository{pool: pool}
}

func (r *PostgresAttemptRepository) CreateAttempt(ctx context.Context, attempt QuizAttempt) (QuizAttempt, error) {
	id, err := newID("attempt")
	if err != nil {
		return QuizAttempt{}, err
	}
	attempt.ID = id

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return QuizAttempt{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO quiz_attempts (id, user_id, topic_id, score, total, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		attempt.ID, attempt.UserID, attempt.TopicID, attempt.Score, attempt.Total, attempt.CreatedAt)
	if err != nil {
		return QuizAttempt{}, err
	}

	for i, a := range attempt.Answers {
		var qid any
		if a.QuestionID != "" {
			qid = a.QuestionID
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO quiz_attempt_answers
				(attempt_id, position, question_id, question_text, selected_option_text, correct_option_text, is_correct)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			attempt.ID, i+1, qid, a.QuestionText, a.SelectedOptionText, a.CorrectOptionText, a.IsCorrect)
		if err != nil {
			return QuizAttempt{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return QuizAttempt{}, err
	}
	attempt.Answers = cloneAttemptAnswers(attempt.Answers)
	return attempt, nil
}

func (r *PostgresAttemptRepository) ListAttemptsByUserID(ctx context.Context, userID string) ([]QuizAttempt, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.user_id, a.topic_id, t.slug, a.score, a.total, a.created_at
		 FROM quiz_attempts a
		 JOIN topics t ON t.id = a.topic_id
		 WHERE a.user_id = $1
		 ORDER BY a.created_at DESC, a.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attempts := make([]QuizAttempt, 0)
	ids := make([]string, 0)
	for rows.Next() {
		var a QuizAttempt
		if err := rows.Scan(&a.ID, &a.UserID, &a.TopicID, &a.TopicSlug, &a.Score, &a.Total, &a.CreatedAt); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
		ids = append(ids, a.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(attempts) == 0 {
		return attempts, nil
	}
	answersByAttempt, err := r.loadAnswers(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range attempts {
		attempts[i].Answers = answersByAttempt[attempts[i].ID]
	}
	return attempts, nil
}

func (r *PostgresAttemptRepository) GetAttemptByID(ctx context.Context, id string) (QuizAttempt, error) {
	var a QuizAttempt
	err := r.pool.QueryRow(ctx,
		`SELECT a.id, a.user_id, a.topic_id, t.slug, a.score, a.total, a.created_at
		 FROM quiz_attempts a
		 JOIN topics t ON t.id = a.topic_id
		 WHERE a.id = $1`, id).
		Scan(&a.ID, &a.UserID, &a.TopicID, &a.TopicSlug, &a.Score, &a.Total, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuizAttempt{}, ErrAttemptNotFound
	}
	if err != nil {
		return QuizAttempt{}, err
	}
	answersByAttempt, err := r.loadAnswers(ctx, []string{id})
	if err != nil {
		return QuizAttempt{}, err
	}
	a.Answers = answersByAttempt[id]
	return a, nil
}

func (r *PostgresAttemptRepository) loadAnswers(ctx context.Context, attemptIDs []string) (map[string][]QuizAttemptAnswer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT attempt_id, question_id, question_text, selected_option_text, correct_option_text, is_correct
		 FROM quiz_attempt_answers
		 WHERE attempt_id = ANY($1)
		 ORDER BY attempt_id, position`, attemptIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]QuizAttemptAnswer, len(attemptIDs))
	for _, id := range attemptIDs {
		out[id] = make([]QuizAttemptAnswer, 0)
	}
	for rows.Next() {
		var attemptID string
		var a QuizAttemptAnswer
		var qid *string
		if err := rows.Scan(&attemptID, &qid, &a.QuestionText, &a.SelectedOptionText, &a.CorrectOptionText, &a.IsCorrect); err != nil {
			return nil, err
		}
		if qid != nil {
			a.QuestionID = *qid
		}
		out[attemptID] = append(out[attemptID], a)
	}
	return out, rows.Err()
}
