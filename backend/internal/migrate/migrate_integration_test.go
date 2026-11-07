//go:build integration

package migrate

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
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
	return url
}

func openTestConn(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testDBURL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })
	return conn
}

func resetSchema(t *testing.T, conn *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	// Drop in dependency order so repeated test runs start clean.
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
		if _, err := conn.Exec(ctx, s); err != nil {
			t.Fatalf("reset schema (%s): %v", s, err)
		}
	}
}

func TestApply_ContentTables(t *testing.T) {
	ctx := context.Background()
	conn := openTestConn(t)
	resetSchema(t, conn)

	fsys := os.DirFS("../../migrations")

	if err := Apply(ctx, conn, fsys); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Tables exist with status columns.
	tables := []string{"topics", "lessons", "questions", "answer_options"}
	for _, table := range tables {
		var hasStatus bool
		err := conn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = $1 AND column_name = 'status'
			)`, table).Scan(&hasStatus)
		if err != nil {
			t.Fatalf("check status column on %s: %v", table, err)
		}
		if !hasStatus {
			t.Fatalf("table %s missing status column", table)
		}
	}

	// Migrations recorded.
	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 recorded migration, got %d", count)
	}
	firstCount := count

	// Applying twice leaves same schema and does not duplicate records.
	if err := Apply(ctx, conn, fsys); err != nil {
		t.Fatalf("apply second time: %v", err)
	}
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations after second apply: %v", err)
	}
	if count != firstCount {
		t.Fatalf("expected %d recorded migrations after second apply, got %d", firstCount, count)
	}
}

func TestApply_Constraints(t *testing.T) {
	ctx := context.Background()
	conn := openTestConn(t)
	resetSchema(t, conn)

	if err := Apply(ctx, conn, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Topic slug unique.
	_, err := conn.Exec(ctx, `INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t1','s1','T','D',1,'draft')`)
	if err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t2','s1','T2','D',2,'draft')`)
	if err == nil {
		t.Fatal("expected unique slug violation")
	}

	// Lesson position unique per topic.
	_, err = conn.Exec(ctx, `INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l1','t1','L','B',1,'draft')`)
	if err != nil {
		t.Fatalf("insert lesson: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l2','t1','L2','B',1,'draft')`)
	if err == nil {
		t.Fatal("expected unique (topic_id, position) violation for lessons")
	}
	// Same position in a different topic is allowed.
	_, err = conn.Exec(ctx, `INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t3','s3','T3','D',3,'draft')`)
	if err != nil {
		t.Fatalf("insert topic t3: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l3','t3','L3','B',1,'draft')`)
	if err != nil {
		t.Fatalf("same position in different topic should be allowed: %v", err)
	}

	// Question fields from section 9; answer_options position unique per question.
	_, err = conn.Exec(ctx, `INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ('q1','t1','Q?','E','draft')`)
	if err != nil {
		t.Fatalf("insert question: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ('o1','q1','A',1,true,'draft')`)
	if err != nil {
		t.Fatalf("insert option: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO answer_options (id, question_id, text, position, is_correct, status) VALUES ('o2','q1','B',1,false,'draft')`)
	if err == nil {
		t.Fatal("expected unique (question_id, position) violation for answer_options")
	}
}

func TestApply_AuthProgressAttemptTables(t *testing.T) {
	ctx := context.Background()
	conn := openTestConn(t)
	resetSchema(t, conn)

	fsys := os.DirFS("../../migrations")
	if err := Apply(ctx, conn, fsys); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// users: unique email, password hash, is_maintainer default false, created at
	_, err := conn.Exec(ctx, `INSERT INTO users (id, email, password_hash, created_at) VALUES ('u1','a@example.com','hash', NOW())`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var isMaintainer bool
	if err := conn.QueryRow(ctx, `SELECT is_maintainer FROM users WHERE id = 'u1'`).Scan(&isMaintainer); err != nil {
		t.Fatalf("select is_maintainer: %v", err)
	}
	if isMaintainer {
		t.Fatal("is_maintainer should default to false")
	}
	_, err = conn.Exec(ctx, `INSERT INTO users (id, email, password_hash, created_at) VALUES ('u2','a@example.com','hash2', NOW())`)
	if err == nil {
		t.Fatal("expected unique email violation")
	}

	// sessions: id, user_id, created_at, expires_at, indexes on user_id and expires_at
	_, err = conn.Exec(ctx, `INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES ('s1','u1', NOW(), NOW() + INTERVAL '7 days')`)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	for _, idx := range []string{"sessions_user_id_idx", "sessions_expires_at_idx"} {
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'sessions' AND indexname = $1)`, idx).Scan(&exists); err != nil {
			t.Fatalf("check index %s: %v", idx, err)
		}
		if !exists {
			t.Fatalf("missing index %s", idx)
		}
	}

	// lesson_progress: (user_id, lesson_id) key and completion timestamp
	_, err = conn.Exec(ctx, `INSERT INTO topics (id, slug, title, description, position, status) VALUES ('t1','slug1','T','D',1,'published')`)
	if err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO lessons (id, topic_id, title, body, position, status) VALUES ('l1','t1','L','B',1,'published')`)
	if err != nil {
		t.Fatalf("insert lesson: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO lesson_progress (user_id, lesson_id, completed_at) VALUES ('u1','l1', NOW())`)
	if err != nil {
		t.Fatalf("insert lesson_progress: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO lesson_progress (user_id, lesson_id, completed_at) VALUES ('u1','l1', NOW())`)
	if err == nil {
		t.Fatal("expected duplicate lesson_progress key violation")
	}

	// quiz_attempts after users; quiz_attempt_answers with nullable question ref and snapshots
	_, err = conn.Exec(ctx, `INSERT INTO questions (id, topic_id, text, explanation, status) VALUES ('q1','t1','Q?','E','published')`)
	if err != nil {
		t.Fatalf("insert question: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO quiz_attempts (id, user_id, topic_id, score, total, created_at) VALUES ('a1','u1','t1',1,1, NOW())`)
	if err != nil {
		t.Fatalf("insert quiz_attempt: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO quiz_attempt_answers (attempt_id, question_id, question_text, selected_option_text, correct_option_text, is_correct) VALUES ('a1','q1','Q?','A','A',true)`)
	if err != nil {
		t.Fatalf("insert quiz_attempt_answer: %v", err)
	}
	// nullable question reference and empty selected answer allowed
	_, err = conn.Exec(ctx, `INSERT INTO quiz_attempt_answers (attempt_id, question_id, question_text, selected_option_text, correct_option_text, is_correct) VALUES ('a1',NULL,'Gone?','','B',false)`)
	if err != nil {
		t.Fatalf("insert answer with null question_id: %v", err)
	}

	// migration ordering: quiz_attempts migration version sorts after users (both in 000002, and file after 000001)
	// verify FK prevents orphan attempt
	_, err = conn.Exec(ctx, `INSERT INTO quiz_attempts (id, user_id, topic_id, score, total, created_at) VALUES ('a2','nope','t1',0,1, NOW())`)
	if err == nil {
		t.Fatal("expected FK violation for missing user")
	}
}
