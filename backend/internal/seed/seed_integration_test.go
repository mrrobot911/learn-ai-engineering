//go:build integration

package seed_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
	"github.com/mrrobot911/learn-ai-engineering/internal/seed"
	seeddata "github.com/mrrobot911/learn-ai-engineering/seed"
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
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}
	dbName := cfg.ConnConfig.Database
	if dbName == "learn_ai_engineering" || !strings.Contains(strings.ToLower(dbName), "test") {
		t.Fatalf("refusing to run integration tests against non-test database %q (set TEST_DATABASE_URL to an isolated test database)", dbName)
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

func resetTables(t *testing.T, pool *pgxpool.Pool) {
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

func applyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()
	if err := migrate.Apply(ctx, conn.Conn(), os.DirFS(migrationsDir(t))); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func TestLoad_IsRepeatable(t *testing.T) {
	ctx := context.Background()
	pool := openTestPool(t)
	resetTables(t, pool)
	applyMigrations(t, pool)

	first, err := seed.Load(ctx, pool, seeddata.Files)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if first.Created() == 0 {
		t.Fatal("expected first load to create seed content")
	}

	topics := []string{
		"topic-llms",
		"topic-ai-agents",
		"topic-mcp",
		"topic-rag",
		"topic-embeddings",
		"topic-evaluation",
		"topic-deployment",
	}

	count := func(query string, args ...any) int {
		t.Helper()
		var got int
		if err := pool.QueryRow(ctx, query, args...).Scan(&got); err != nil {
			t.Fatalf("count query %q: %v", query, err)
		}
		return got
	}
	assertCount := func(query string, want int, args ...any) {
		t.Helper()
		if got := count(query, args...); got != want {
			t.Fatalf("count %q = %d, want %d", query, got, want)
		}
	}

	// Assert minimum counts for each of the seven topics after one load.
	type topicCounts struct{ lessons, questions, options int }
	before := make(map[string]topicCounts, len(topics))
	for _, id := range topics {
		assertCount(`SELECT COUNT(*) FROM topics WHERE id = $1`, 1, id)
		lessons := count(`SELECT COUNT(*) FROM lessons WHERE topic_id = $1 AND status = 'published'`, id)
		questions := count(`SELECT COUNT(*) FROM questions WHERE topic_id = $1 AND status = 'published'`, id)
		options := count(`SELECT COUNT(*) FROM answer_options ao JOIN questions q ON q.id = ao.question_id WHERE q.topic_id = $1`, id)
		if lessons < 1 || questions < 6 {
			t.Fatalf("topic %s counts: lessons=%d questions=%d, want >=1 lesson and >=6 questions", id, lessons, questions)
		}
		before[id] = topicCounts{lessons: lessons, questions: questions, options: options}
	}

	second, err := seed.Load(ctx, pool, seeddata.Files)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if second.Created() != 0 {
		t.Fatalf("expected second load to create nothing, got %+v", second)
	}

	// Assert database counts unchanged after second load.
	for _, id := range topics {
		want := before[id]
		assertCount(`SELECT COUNT(*) FROM lessons WHERE topic_id = $1`, want.lessons, id)
		assertCount(`SELECT COUNT(*) FROM questions WHERE topic_id = $1`, want.questions, id)
		assertCount(`SELECT COUNT(*) FROM answer_options ao JOIN questions q ON q.id = ao.question_id WHERE q.topic_id = $1`, want.options, id)
	}
}
