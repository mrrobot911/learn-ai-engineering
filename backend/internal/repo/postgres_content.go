package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresContentRepository is a PostgreSQL implementation of ContentRepository.
type PostgresContentRepository struct {
	pool *pgxpool.Pool
}

var _ ContentRepository = (*PostgresContentRepository)(nil)

// NewPostgresContentRepository creates a content repository backed by pool.
func NewPostgresContentRepository(pool *pgxpool.Pool) *PostgresContentRepository {
	return &PostgresContentRepository{pool: pool}
}

func newID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "-" + hex.EncodeToString(b), nil
}

func (r *PostgresContentRepository) ListPublishedTopics(ctx context.Context) ([]Topic, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, slug, title, description, position, status
		 FROM topics WHERE status = $1 ORDER BY position`, StatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topics := make([]Topic, 0)
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.ID, &t.Slug, &t.Title, &t.Description, &t.Position, &t.Status); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func (r *PostgresContentRepository) GetPublishedTopicBySlug(ctx context.Context, slug string) (Topic, error) {
	var t Topic
	err := r.pool.QueryRow(ctx,
		`SELECT id, slug, title, description, position, status
		 FROM topics WHERE slug = $1 AND status = $2`, slug, StatusPublished).
		Scan(&t.ID, &t.Slug, &t.Title, &t.Description, &t.Position, &t.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Topic{}, ErrNotFound
	}
	return t, err
}

func (r *PostgresContentRepository) GetTopicBySlug(ctx context.Context, slug string) (Topic, error) {
	var t Topic
	err := r.pool.QueryRow(ctx,
		`SELECT id, slug, title, description, position, status
		 FROM topics WHERE slug = $1`, slug).
		Scan(&t.ID, &t.Slug, &t.Title, &t.Description, &t.Position, &t.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Topic{}, ErrNotFound
	}
	return t, err
}

func (r *PostgresContentRepository) ListPublishedLessonsByTopicID(ctx context.Context, topicID string) ([]Lesson, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, topic_id, title, body, position, status
		 FROM lessons WHERE topic_id = $1 AND status = $2 ORDER BY position`, topicID, StatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := make([]Lesson, 0)
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(&l.ID, &l.TopicID, &l.Title, &l.Body, &l.Position, &l.Status); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *PostgresContentRepository) GetPublishedLessonByID(ctx context.Context, id string) (Lesson, error) {
	var l Lesson
	err := r.pool.QueryRow(ctx,
		`SELECT id, topic_id, title, body, position, status
		 FROM lessons WHERE id = $1 AND status = $2`, id, StatusPublished).
		Scan(&l.ID, &l.TopicID, &l.Title, &l.Body, &l.Position, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	return l, err
}

func (r *PostgresContentRepository) GetLessonByID(ctx context.Context, id string) (Lesson, error) {
	var l Lesson
	err := r.pool.QueryRow(ctx,
		`SELECT id, topic_id, title, body, position, status
		 FROM lessons WHERE id = $1`, id).
		Scan(&l.ID, &l.TopicID, &l.Title, &l.Body, &l.Position, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	return l, err
}

func (r *PostgresContentRepository) CreateDraftLesson(ctx context.Context, topicID string, lesson Lesson) (Lesson, error) {
	id, err := newID("lesson")
	if err != nil {
		return Lesson{}, err
	}
	lesson.ID = id
	lesson.TopicID = topicID
	lesson.Status = StatusDraft
	_, err = r.pool.Exec(ctx,
		`INSERT INTO lessons (id, topic_id, title, body, position, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		lesson.ID, lesson.TopicID, lesson.Title, lesson.Body, lesson.Position, lesson.Status)
	return lesson, err
}

func (r *PostgresContentRepository) UpdateDraftLesson(ctx context.Context, lesson Lesson) (Lesson, error) {
	lesson.Status = StatusDraft
	tag, err := r.pool.Exec(ctx,
		`UPDATE lessons SET topic_id = $2, title = $3, body = $4, position = $5, status = $6
		 WHERE id = $1 AND status = $7`,
		lesson.ID, lesson.TopicID, lesson.Title, lesson.Body, lesson.Position, lesson.Status, StatusDraft)
	if err != nil {
		return Lesson{}, err
	}
	if tag.RowsAffected() == 0 {
		return Lesson{}, ErrNotFound
	}
	return lesson, nil
}

func (r *PostgresContentRepository) PublishLesson(ctx context.Context, topicID, lessonID string) (Lesson, error) {
	var l Lesson
	err := r.pool.QueryRow(ctx,
		`UPDATE lessons SET status = $3
		 WHERE id = $1 AND topic_id = $2
		 RETURNING id, topic_id, title, body, position, status`,
		lessonID, topicID, StatusPublished).
		Scan(&l.ID, &l.TopicID, &l.Title, &l.Body, &l.Position, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	return l, err
}

func (r *PostgresContentRepository) UnpublishLesson(ctx context.Context, topicID, lessonID string) (Lesson, error) {
	var l Lesson
	err := r.pool.QueryRow(ctx,
		`UPDATE lessons SET status = $3
		 WHERE id = $1 AND topic_id = $2
		 RETURNING id, topic_id, title, body, position, status`,
		lessonID, topicID, StatusDraft).
		Scan(&l.ID, &l.TopicID, &l.Title, &l.Body, &l.Position, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	return l, err
}

func (r *PostgresContentRepository) ListPublishedQuestionsByTopicID(ctx context.Context, topicID string) ([]Question, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, topic_id, text, explanation, status
		 FROM questions WHERE topic_id = $1 AND status = $2`, topicID, StatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]Question, 0)
	ids := make([]string, 0)
	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.TopicID, &q.Text, &q.Explanation, &q.Status); err != nil {
			return nil, err
		}
		questions = append(questions, q)
		ids = append(ids, q.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return questions, nil
	}
	optsByQ, err := r.loadOptions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range questions {
		questions[i].Options = optsByQ[questions[i].ID]
	}
	return questions, nil
}

func (r *PostgresContentRepository) GetQuestionByID(ctx context.Context, id string) (Question, error) {
	var q Question
	err := r.pool.QueryRow(ctx,
		`SELECT id, topic_id, text, explanation, status
		 FROM questions WHERE id = $1`, id).
		Scan(&q.ID, &q.TopicID, &q.Text, &q.Explanation, &q.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	optsByQ, err := r.loadOptions(ctx, []string{id})
	if err != nil {
		return Question{}, err
	}
	q.Options = optsByQ[id]
	return q, nil
}

func (r *PostgresContentRepository) loadOptions(ctx context.Context, questionIDs []string) (map[string][]Option, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, question_id, text, position, is_correct
		 FROM answer_options WHERE question_id = ANY($1) ORDER BY question_id, position`, questionIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]Option, len(questionIDs))
	for _, id := range questionIDs {
		out[id] = make([]Option, 0)
	}
	for rows.Next() {
		var o Option
		var qid string
		if err := rows.Scan(&o.ID, &qid, &o.Text, &o.Position, &o.IsCorrect); err != nil {
			return nil, err
		}
		out[qid] = append(out[qid], o)
	}
	return out, rows.Err()
}

func (r *PostgresContentRepository) CreateDraftQuestion(ctx context.Context, topicID string, question Question) (Question, error) {
	id, err := newID("question")
	if err != nil {
		return Question{}, err
	}
	question.ID = id
	question.TopicID = topicID
	question.Status = StatusDraft
	question.Options = assignOptionIDs(question.Options)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`INSERT INTO questions (id, topic_id, text, explanation, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		question.ID, question.TopicID, question.Text, question.Explanation, question.Status); err != nil {
		return Question{}, err
	}
	if err := insertOptions(ctx, tx, question.ID, question.Options); err != nil {
		return Question{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Question{}, err
	}
	return question, nil
}

func (r *PostgresContentRepository) UpdateDraftQuestion(ctx context.Context, question Question) (Question, error) {
	question.Status = StatusDraft
	question.Options = assignOptionIDs(question.Options)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE questions SET topic_id = $2, text = $3, explanation = $4, status = $5
		 WHERE id = $1 AND status = $6`,
		question.ID, question.TopicID, question.Text, question.Explanation, question.Status, StatusDraft)
	if err != nil {
		return Question{}, err
	}
	if tag.RowsAffected() == 0 {
		return Question{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM answer_options WHERE question_id = $1`, question.ID); err != nil {
		return Question{}, err
	}
	if err := insertOptions(ctx, tx, question.ID, question.Options); err != nil {
		return Question{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Question{}, err
	}
	return question, nil
}

func (r *PostgresContentRepository) PublishQuestion(ctx context.Context, topicID, questionID string) (Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var q Question
	err = tx.QueryRow(ctx,
		`UPDATE questions SET status = $3
		 WHERE id = $1 AND topic_id = $2
		 RETURNING id, topic_id, text, explanation, status`,
		questionID, topicID, StatusPublished).
		Scan(&q.ID, &q.TopicID, &q.Text, &q.Explanation, &q.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	opts, err := loadOptionsTx(ctx, tx, questionID)
	if err != nil {
		return Question{}, err
	}
	q.Options = opts
	if err := tx.Commit(ctx); err != nil {
		return Question{}, err
	}
	return q, nil
}

func (r *PostgresContentRepository) UnpublishQuestion(ctx context.Context, topicID, questionID string) (Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var q Question
	err = tx.QueryRow(ctx,
		`UPDATE questions SET status = $3
		 WHERE id = $1 AND topic_id = $2
		 RETURNING id, topic_id, text, explanation, status`,
		questionID, topicID, StatusDraft).
		Scan(&q.ID, &q.TopicID, &q.Text, &q.Explanation, &q.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	opts, err := loadOptionsTx(ctx, tx, questionID)
	if err != nil {
		return Question{}, err
	}
	q.Options = opts
	if err := tx.Commit(ctx); err != nil {
		return Question{}, err
	}
	return q, nil
}

func loadOptionsTx(ctx context.Context, tx pgx.Tx, questionID string) ([]Option, error) {
	rows, err := tx.Query(ctx,
		`SELECT id, question_id, text, position, is_correct
		 FROM answer_options WHERE question_id = $1 ORDER BY position`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Option, 0)
	for rows.Next() {
		var o Option
		var qid string
		if err := rows.Scan(&o.ID, &qid, &o.Text, &o.Position, &o.IsCorrect); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func insertOptions(ctx context.Context, tx pgx.Tx, questionID string, options []Option) error {
	for _, o := range options {
		// answer_options has status column per T21a migration (NOT NULL, checked);
		// options follow their question's visibility in queries.
		if _, err := tx.Exec(ctx,
			`INSERT INTO answer_options (id, question_id, text, position, is_correct, status)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			o.ID, questionID, o.Text, o.Position, o.IsCorrect, StatusDraft); err != nil {
			return err
		}
	}
	return nil
}

func assignOptionIDs(options []Option) []Option {
	out := make([]Option, len(options))
	for i, o := range options {
		id, err := newID("option")
		if err != nil {
			// fallback: deterministic enough for test contexts; crypto/rand errors are rare
			id = fmt.Sprintf("option-%d", i+1)
		}
		out[i] = o
		out[i].ID = id
		if out[i].Position == 0 {
			out[i].Position = i + 1
		}
	}
	return out
}
