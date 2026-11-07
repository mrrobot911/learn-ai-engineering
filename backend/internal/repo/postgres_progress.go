package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresProgressRepository is a PostgreSQL implementation of ProgressRepository.
type PostgresProgressRepository struct {
	pool *pgxpool.Pool
}

var _ ProgressRepository = (*PostgresProgressRepository)(nil)

// NewPostgresProgressRepository creates a progress repository backed by pool.
func NewPostgresProgressRepository(pool *pgxpool.Pool) *PostgresProgressRepository {
	return &PostgresProgressRepository{pool: pool}
}

func (r *PostgresProgressRepository) ListLessonProgress(ctx context.Context, userID string) ([]LessonProgress, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, lesson_id, completed_at
		 FROM lesson_progress WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progress := make([]LessonProgress, 0)
	for rows.Next() {
		var p LessonProgress
		if err := rows.Scan(&p.UserID, &p.LessonID, &p.CompletedAt); err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}
	return progress, rows.Err()
}

func (r *PostgresProgressRepository) MarkLessonComplete(ctx context.Context, userID, lessonID string, completedAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO lesson_progress (user_id, lesson_id, completed_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, lesson_id) DO NOTHING`,
		userID, lessonID, completedAt)
	return err
}

func (r *PostgresProgressRepository) UnmarkLessonComplete(ctx context.Context, userID, lessonID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM lesson_progress WHERE user_id = $1 AND lesson_id = $2`,
		userID, lessonID)
	return err
}
