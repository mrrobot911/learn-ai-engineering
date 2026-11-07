//go:build integration

package repo

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
)

func testDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL required for integration tests")
	}
	// AC4 (T21b): integration tests must never point at the development database.
	// Refuse known dev DB names and require a test database name.
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}
	dbName := cfg.ConnConfig.Database
	if dbName == "learn_ai_engineering" || !strings.Contains(strings.ToLower(dbName), "test") {
		t.Fatalf("refusing to run integration tests against non-test database %q (set TEST_DATABASE_URL to an isolated test database, e.g. learn_ai_engineering_test)", dbName)
	}
	return url
}

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDBURL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

func resetContentTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DROP TABLE IF EXISTS quiz_attempt_answers`,
		`DROP TABLE IF EXISTS quiz_attempts`,
		`DROP TABLE IF EXISTS lesson_progress`,
		`DROP TABLE IF EXISTS sessions`,
		`DROP TABLE IF EXISTS users`,
		`DROP TABLE IF EXISTS answer_options`,
		`DROP TABLE IF EXISTS questions`,
		`DROP TABLE IF EXISTS lessons`,
		`DROP TABLE IF EXISTS topics`,
		`DROP TABLE IF EXISTS schema_migrations`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			t.Fatalf("reset schema (%s): %v", s, err)
		}
	}
}

func TestPostgresContentRepository_PublishedFilterAndPositionOrdering(t *testing.T) {
	ctx := context.Background()
	pool := openTestPool(t)
	resetContentTables(t, pool)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	if err := migrate.Apply(ctx, conn.Conn(), os.DirFS(migrationsDir(t))); err != nil {
		conn.Release()
		t.Fatalf("apply migrations: %v", err)
	}
	conn.Release()

	r := NewPostgresContentRepository(pool)

	// Topics: mixed status, out-of-order positions.
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("exec: %v", err)
		}
	}
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t-draft','draft','Draft','',1,'draft')`)
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t-pub-2','pub-2','Pub 2','',3,'published')`)
	exec(`INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t-pub-1','pub-1','Pub 1','',1,'published')`)

	topics, err := r.ListPublishedTopics(ctx)
	if err != nil {
		t.Fatalf("ListPublishedTopics: %v", err)
	}
	if len(topics) != 2 {
		t.Fatalf("expected 2 published topics, got %d", len(topics))
	}
	if topics[0].Slug != "pub-1" || topics[1].Slug != "pub-2" {
		t.Fatalf("topics not in position order: got %q, %q", topics[0].Slug, topics[1].Slug)
	}

	// Lessons: mixed status under published topic, out-of-order positions.
	exec(`INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l-draft','t-pub-1','Draft','',1,'draft')`)
	exec(`INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l-2','t-pub-1','L2','',5,'published')`)
	exec(`INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l-1','t-pub-1','L1','',2,'published')`)

	lessons, err := r.ListPublishedLessonsByTopicID(ctx, "t-pub-1")
	if err != nil {
		t.Fatalf("ListPublishedLessonsByTopicID: %v", err)
	}
	if len(lessons) != 2 {
		t.Fatalf("expected 2 published lessons, got %d", len(lessons))
	}
	if lessons[0].ID != "l-1" || lessons[1].ID != "l-2" {
		t.Fatalf("lessons not in position order: got %q, %q", lessons[0].ID, lessons[1].ID)
	}

	// Questions: mixed status; options out-of-order positions.
	exec(`INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ('q-draft','t-pub-1','Q?','', 'draft')`)
	exec(`INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ('q-pub','t-pub-1','Q pub?','E','published')`)
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ('o-3','q-pub','C',3,false,'published')`)
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ('o-1','q-pub','A',1,true,'published')`)
	exec(`INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ('o-2','q-pub','B',2,false,'published')`)

	questions, err := r.ListPublishedQuestionsByTopicID(ctx, "t-pub-1")
	if err != nil {
		t.Fatalf("ListPublishedQuestionsByTopicID: %v", err)
	}
	if len(questions) != 1 || questions[0].ID != "q-pub" {
		t.Fatalf("expected 1 published question q-pub, got %+v", questions)
	}
	if len(questions[0].Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(questions[0].Options))
	}
	if questions[0].Options[0].ID != "o-1" || questions[0].Options[1].ID != "o-2" || questions[0].Options[2].ID != "o-3" {
		t.Fatalf("options not in position order: got %q, %q, %q",
			questions[0].Options[0].ID, questions[0].Options[1].ID, questions[0].Options[2].ID)
	}
}
