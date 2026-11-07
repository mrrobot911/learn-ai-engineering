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

	// Migration recorded.
	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 recorded migration, got %d", count)
	}

	// Applying twice leaves same schema and does not duplicate records.
	if err := Apply(ctx, conn, fsys); err != nil {
		t.Fatalf("apply second time: %v", err)
	}
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations after second apply: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 recorded migration after second apply, got %d", count)
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
