package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

func TestMarkLessonComplete(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	progressService := service.NewProgressService(progressRepo, contentRepo)
	handler := NewProgressHandler(progressService)

	mux := http.NewServeMux()
	mux.Handle("PUT /me/progress/lessons/{lessonId}", http.HandlerFunc(handler.MarkLessonComplete))

	// Test 1: Successful mark
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/me/progress/lessons/lesson1", nil)
	authSession := service.AuthSession{
		User: repo.User{ID: "user1"},
	}
	req = req.WithContext(session.With(req.Context(), authSession))
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("MarkLessonComplete() status = %v, want %v", w.Code, http.StatusOK)
	}

	var progress []string
	if err := json.Unmarshal(w.Body.Bytes(), &progress); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("MarkLessonComplete() progress = %v, want [lesson1]", progress)
	}

	// Test 2: Idempotent - mark again
	w = httptest.NewRecorder()
	handler.MarkLessonComplete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("MarkLessonComplete() idempotent status = %v, want %v", w.Code, http.StatusOK)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &progress); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("MarkLessonComplete() idempotent progress = %v, want [lesson1]", progress)
	}

	// Test 3: No valid session
	req = httptest.NewRequest("PUT", "/me/progress/lessons/lesson1", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	t.Logf("Status: %d", w.Code)
	t.Logf("Body: %s", w.Body.String())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("MarkLessonComplete() unauthenticated status = %v, want %v", w.Code, http.StatusUnauthorized)
	}

	// Test 4: Non-existent lesson
	contentRepo = repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{}, []repo.Question{}) // Clear lessons
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/me/progress/lessons/nonexistent", nil)
	authSession = service.AuthSession{User: repo.User{ID: "user1"}}
	req = req.WithContext(session.With(req.Context(), authSession))

	mux.ServeHTTP(w, req)

	t.Logf("Status: %d", w.Code)
	t.Logf("Body: %s", w.Body.String())

	if w.Code != http.StatusNotFound {
		t.Errorf("MarkLessonComplete() not found status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestUnmarkLessonComplete(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	progressService := service.NewProgressService(progressRepo, contentRepo)
	handler := NewProgressHandler(progressService)

	mux := http.NewServeMux()
	mux.Handle("DELETE /me/progress/lessons/{lessonId}", http.HandlerFunc(handler.UnmarkLessonComplete))

	// Test 1: Successful unmark
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/me/progress/lessons/lesson1", nil)
	authSession := service.AuthSession{User: repo.User{ID: "user1"}}
	req = req.WithContext(session.With(req.Context(), authSession))
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UnmarkLessonComplete() status = %v, want %v", w.Code, http.StatusOK)
	}

	var progress []string
	if err := json.Unmarshal(w.Body.Bytes(), &progress); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if len(progress) != 0 {
		t.Errorf("UnmarkLessonComplete() progress = %v, want []", progress)
	}

	// Test 2: Safe to repeat
	w = httptest.NewRecorder()
	handler.UnmarkLessonComplete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UnmarkLessonComplete() safe to repeat status = %v, want %v", w.Code, http.StatusOK)
	}

	// Test 3: No valid session
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/me/progress/lessons/lesson1", nil)

	mux.ServeHTTP(w, req)

	t.Logf("Status: %d", w.Code)
	t.Logf("Body: %s", w.Body.String())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UnmarkLessonComplete() unauthenticated status = %v, want %v", w.Code, http.StatusUnauthorized)
	}

	// Test 4: Non-existent lesson
	contentRepo = repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{}, []repo.Question{}) // Clear lessons
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/me/progress/lessons/nonexistent", nil)
	authSession = service.AuthSession{User: repo.User{ID: "user1"}}
	req = req.WithContext(session.With(req.Context(), authSession))

	mux.ServeHTTP(w, req)

	t.Logf("Status: %d", w.Code)
	t.Logf("Body: %s", w.Body.String())

	if w.Code != http.StatusNotFound {
		t.Errorf("UnmarkLessonComplete() not found status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestGetMyProgress(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	progressService := service.NewProgressService(progressRepo, contentRepo)
	handler := NewProgressHandler(progressService)

	mux := http.NewServeMux()
	mux.Handle("GET /me/progress", http.HandlerFunc(handler.GetMyProgress))

	// Test 1: User with no progress
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/me/progress", nil)
	authSession := service.AuthSession{User: repo.User{ID: "user1"}}
	req = req.WithContext(session.With(req.Context(), authSession))
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMyProgress() status = %v, want %v", w.Code, http.StatusOK)
	}

	var progress []string
	if err := json.Unmarshal(w.Body.Bytes(), &progress); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if len(progress) != 0 {
		t.Errorf("GetMyProgress() progress = %v, want []", progress)
	}

	// Test 2: User with progress
	progressService.MarkLessonComplete(context.Background(), "user1", "lesson1")

	w = httptest.NewRecorder()
	handler.GetMyProgress(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMyProgress() status = %v, want %v", w.Code, http.StatusOK)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &progress); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("GetMyProgress() progress = %v, want [lesson1]", progress)
	}

	// Test 3: No valid session
	req = httptest.NewRequest("GET", "/me/progress", nil)
	w = httptest.NewRecorder()

	handler.GetMyProgress(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetMyProgress() unauthenticated status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestOwnershipEnforcement(t *testing.T) {
	progressRepo := repo.NewMemoryProgressRepository()
	contentRepo := repo.NewMemoryContentRepository([]repo.Topic{}, []repo.Lesson{
		{ID: "lesson1", TopicID: "topic1", Status: repo.StatusPublished},
	}, []repo.Question{})

	progressService := service.NewProgressService(progressRepo, contentRepo)
	handler := NewProgressHandler(progressService)

	mux := http.NewServeMux()
	mux.Handle("PUT /me/progress/lessons/{lessonId}", http.HandlerFunc(handler.MarkLessonComplete))
	mux.Handle("GET /me/progress", http.HandlerFunc(handler.GetMyProgress))

	// Create two users
	user1 := service.AuthSession{User: repo.User{ID: "user1"}}
	user2 := service.AuthSession{User: repo.User{ID: "user2"}}

	// User1 marks a lesson
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("PUT", "/me/progress/lessons/lesson1", nil)
	req1 = req1.WithContext(session.With(req1.Context(), user1))
	mux.ServeHTTP(w1, req1)

	// Verify user1 can see their progress
	var progress []string
	json.Unmarshal(w1.Body.Bytes(), &progress)
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("User1 progress = %v, want [lesson1]", progress)
	}

	// Try to mark with user2's context (should work as separate users)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", "/me/progress/lessons/lesson1", nil)
	req2 = req2.WithContext(session.With(req2.Context(), user2))
	mux.ServeHTTP(w2, req2)

	// Both users should see their own progress
	json.Unmarshal(w2.Body.Bytes(), &progress)
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("User2 progress = %v, want [lesson1]", progress)
	}

	// User1 should still see only their own progress
	w1 = httptest.NewRecorder()
	req1 = httptest.NewRequest("GET", "/me/progress", nil)
	req1 = req1.WithContext(session.With(req1.Context(), user1))
	mux.ServeHTTP(w1, req1)
	json.Unmarshal(w1.Body.Bytes(), &progress)
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("User1 should see only their own progress = %v, want [lesson1]", progress)
	}

	// User2 should see only their own progress
	w2 = httptest.NewRecorder()
	req2 = httptest.NewRequest("GET", "/me/progress", nil)
	req2 = req2.WithContext(session.With(req2.Context(), user2))
	mux.ServeHTTP(w2, req2)
	json.Unmarshal(w2.Body.Bytes(), &progress)
	if len(progress) != 1 || progress[0] != "lesson1" {
		t.Errorf("User2 should see only their own progress = %v, want [lesson1]", progress)
	}
}
