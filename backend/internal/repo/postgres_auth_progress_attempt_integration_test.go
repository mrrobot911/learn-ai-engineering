//go:build integration

package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
)

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

func seedTopicAndLesson(t *testing.T, pool *pgxpool.Pool) (topicID, lessonID string) {
	t.Helper()
	ctx := context.Background()
	topicID = "topic-t1"
	lessonID = "lesson-l1"
	_, err := pool.Exec(ctx,
		`INSERT INTO topics (id, slug, title, description, position, status)
		 VALUES ($1, $2, $3, '', 1, $4)`,
		topicID, "t1", "Topic 1", StatusPublished)
	if err != nil {
		t.Fatalf("seed topic: %v", err)
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO lessons (id, topic_id, title, body, position, status)
		 VALUES ($1, $2, $3, '', 1, $4)`,
		lessonID, topicID, "Lesson 1", StatusPublished)
	if err != nil {
		t.Fatalf("seed lesson: %v", err)
	}
	return topicID, lessonID
}

func TestPostgresAuthAndProgress_PersistAcrossRestartAndExpireSession(t *testing.T) {
	ctx := context.Background()
	pool := openTestPool(t)
	resetContentTables(t, pool)
	applyMigrations(t, pool)
	_, lessonID := seedTopicAndLesson(t, pool)

	authRepo := NewPostgresAuthRepository(pool)
	progressRepo := NewPostgresProgressRepository(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Register a user.
	user := User{
		ID:           "user-u1",
		Email:        "user@example.com",
		PasswordHash: "$argon2id$hash-not-exposed",
		IsMaintainer: false,
		CreatedAt:    now,
	}
	if err := authRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	dup := user
	dup.ID = "user-u2"
	if err := authRepo.CreateUser(ctx, dup); err != ErrConflict {
		t.Fatalf("expected duplicate email conflict, got: %v", err)
	}

	// Log in: find user by email and create a session.
	foundUser, err := authRepo.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if foundUser.ID != user.ID || foundUser.PasswordHash == "" {
		t.Fatalf("user lookup: %+v", foundUser)
	}
	session := Session{
		ID:        "session-s1",
		UserID:    user.ID,
		CSRFToken: "csrf-token-1",
		CreatedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	if err := authRepo.CreateSession(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Find session.
	foundSession, err := authRepo.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if foundSession.UserID != user.ID || foundSession.CSRFToken != session.CSRFToken {
		t.Fatalf("session: %+v", foundSession)
	}

	// Mark lesson complete.
	completedAt := now.Add(time.Minute)
	if err := progressRepo.MarkLessonComplete(ctx, user.ID, lessonID, completedAt); err != nil {
		t.Fatalf("mark lesson: %v", err)
	}
	// Idempotent mark keeps original completion time.
	if err := progressRepo.MarkLessonComplete(ctx, user.ID, lessonID, completedAt.Add(time.Hour)); err != nil {
		t.Fatalf("mark lesson again: %v", err)
	}

	// Insert an already-expired session to verify it is refused after restart.
	expired := Session{
		ID:        "session-expired",
		UserID:    user.ID,
		CSRFToken: "csrf-expired",
		CreatedAt: now.Add(-48 * time.Hour),
		ExpiresAt: now.Add(-24 * time.Hour),
	}
	if err := authRepo.CreateSession(ctx, expired); err != nil {
		t.Fatalf("create expired session: %v", err)
	}

	dbURL := testDBURL(t)
	pool.Close()

	// Restart connection: new pool, no reset/migrate.
	pool2, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	defer pool2.Close()

	authRepo2 := NewPostgresAuthRepository(pool2)
	progressRepo2 := NewPostgresProgressRepository(pool2)

	// Read state back after restart.
	gotUser, err := authRepo2.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user after restart: %v", err)
	}
	if gotUser.Email != user.Email {
		t.Fatalf("email after restart: got %q", gotUser.Email)
	}
	gotSession, err := authRepo2.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session after restart: %v", err)
	}
	if gotSession.UserID != user.ID {
		t.Fatalf("session user after restart: got %q", gotSession.UserID)
	}

	progress, err := progressRepo2.ListLessonProgress(ctx, user.ID)
	if err != nil {
		t.Fatalf("list progress after restart: %v", err)
	}
	if len(progress) != 1 || progress[0].LessonID != lessonID {
		t.Fatalf("progress after restart: %+v", progress)
	}
	if !progress[0].CompletedAt.Equal(completedAt) {
		t.Fatalf("completed_at after restart: got %v want %v", progress[0].CompletedAt, completedAt)
	}

	// Expired session refused after restart: data persists and is identifiable as expired.
	expiredGot, err := authRepo2.GetSessionByID(ctx, expired.ID)
	if err != nil {
		t.Fatalf("get expired session after restart: %v", err)
	}
	if time.Now().Before(expiredGot.ExpiresAt) {
		t.Fatalf("expired session should not be valid, expires_at=%v", expiredGot.ExpiresAt)
	}
}

func TestPostgresAttempts_SnapshotSurvivesQuestionChangeAndDelete(t *testing.T) {
	ctx := context.Background()
	pool := openTestPool(t)
	resetContentTables(t, pool)
	applyMigrations(t, pool)
	topicID, _ := seedTopicAndLesson(t, pool)

	// Seed user and question.
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, is_maintainer, created_at)
		 VALUES ($1, $2, $3, false, $4)`,
		"user-u1", "u1@example.com", "hashed-not-a-password", now)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO questions (id, topic_id, text, explanation, status)
		 VALUES ($1, $2, $3, '', $4)`,
		"q1", topicID, "Original question text", StatusPublished)
	if err != nil {
		t.Fatalf("seed question: %v", err)
	}

	attemptRepo := NewPostgresAttemptRepository(pool)

	// Store first attempt with snapshot of original text.
	a1, err := attemptRepo.CreateAttempt(ctx, QuizAttempt{
		UserID:    "user-u1",
		TopicID:   topicID,
		TopicSlug: "t1",
		Score:     1,
		Total:     1,
		CreatedAt: now,
		Answers: []QuizAttemptAnswer{{
			QuestionID:         "q1",
			QuestionText:       "Original question text",
			SelectedOptionText: "A",
			CorrectOptionText:  "A",
			IsCorrect:          true,
		}},
	})
	if err != nil {
		t.Fatalf("create attempt 1: %v", err)
	}
	if a1.ID == "" {
		t.Fatal("attempt 1 missing ID")
	}

	// Change question text in DB (simulates content edit).
	_, err = pool.Exec(ctx, `UPDATE questions SET text = $1 WHERE id = $2`, "Changed question text", "q1")
	if err != nil {
		t.Fatalf("update question: %v", err)
	}

	// Store second attempt using new text (as the application would snapshot now).
	a2, err := attemptRepo.CreateAttempt(ctx, QuizAttempt{
		UserID:    "user-u1",
		TopicID:   topicID,
		TopicSlug: "t1",
		Score:     0,
		Total:     1,
		CreatedAt: now.Add(time.Second),
		Answers: []QuizAttemptAnswer{{
			QuestionID:         "q1",
			QuestionText:       "Changed question text",
			SelectedOptionText: "B",
			CorrectOptionText:  "A",
			IsCorrect:          false,
		}},
	})
	if err != nil {
		t.Fatalf("create attempt 2: %v", err)
	}

	// Read both: stored snapshots unchanged / reflect what was stored.
	got1, err := attemptRepo.GetAttemptByID(ctx, a1.ID)
	if err != nil {
		t.Fatalf("get attempt 1: %v", err)
	}
	got2, err := attemptRepo.GetAttemptByID(ctx, a2.ID)
	if err != nil {
		t.Fatalf("get attempt 2: %v", err)
	}
	if got1.Answers[0].QuestionText != "Original question text" {
		t.Fatalf("attempt 1 snapshot changed: got %q", got1.Answers[0].QuestionText)
	}
	if got2.Answers[0].QuestionText != "Changed question text" {
		t.Fatalf("attempt 2 snapshot: got %q", got2.Answers[0].QuestionText)
	}
	if got1.TopicSlug != "t1" || got2.TopicSlug != "t1" {
		t.Fatalf("topic slug: %q %q", got1.TopicSlug, got2.TopicSlug)
	}

	// List returns both for the user (newest first).
	list, err := attemptRepo.ListAttemptsByUserID(ctx, "user-u1")
	if err != nil {
		t.Fatalf("list attempts: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(list))
	}
	if list[0].ID != a2.ID || list[1].ID != a1.ID {
		t.Fatalf("attempt order: got [%s %s]", list[0].ID, list[1].ID)
	}

	// Deleting a question does not break reading an old attempt (ON DELETE SET NULL).
	_, err = pool.Exec(ctx, `DELETE FROM questions WHERE id = $1`, "q1")
	if err != nil {
		t.Fatalf("delete question: %v", err)
	}
	got1AfterDelete, err := attemptRepo.GetAttemptByID(ctx, a1.ID)
	if err != nil {
		t.Fatalf("get attempt 1 after question delete: %v", err)
	}
	if got1AfterDelete.Answers[0].QuestionText != "Original question text" {
		t.Fatalf("snapshot after question delete: got %q", got1AfterDelete.Answers[0].QuestionText)
	}
	if got1AfterDelete.Answers[0].QuestionID != "" {
		t.Fatalf("question_id should be cleared after delete, got %q", got1AfterDelete.Answers[0].QuestionID)
	}
}
