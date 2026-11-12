package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const indexHTML = "<!doctype html><title>app</title>"

func writeStaticFile(t *testing.T, dir, name, content string) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("create static dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write static file: %v", err)
	}
}

func TestRouterServesAPIUnderPrefix(t *testing.T) {
	router := NewRouter(newFixtureServer(), t.TempDir())

	w := get(t, router, "/api/topics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var topics []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &topics); err != nil {
		t.Fatalf("decode topics: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("expected 2 published topics, got %d", len(topics))
	}
}

func TestRouterServesStaticFile(t *testing.T) {
	dir := t.TempDir()
	writeStaticFile(t, dir, "assets/app.js", "console.log('app')")
	router := NewRouter(http.NewServeMux(), dir)

	w := get(t, router, "/assets/app.js")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "console.log('app')" {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestRouterFallsBackToIndexForUnknownPath(t *testing.T) {
	dir := t.TempDir()
	writeStaticFile(t, dir, "index.html", indexHTML)
	router := NewRouter(http.NewServeMux(), dir)

	w := get(t, router, "/topics/intro")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != indexHTML {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestRouterServesHealthWithoutPrefix(t *testing.T) {
	api := http.NewServeMux()
	api.HandleFunc("GET /health", NewHealthHandler().Health)
	dir := t.TempDir()
	writeStaticFile(t, dir, "index.html", indexHTML)
	router := NewRouter(api, dir)

	w := get(t, router, "/health")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if response["status"] != "ok" {
		t.Errorf("expected status ok, got %s", response["status"])
	}
}

func TestRouterReturnsNotFoundForUnknownAPIPath(t *testing.T) {
	dir := t.TempDir()
	writeStaticFile(t, dir, "index.html", indexHTML)
	router := NewRouter(newFixtureServer(), dir)

	w := get(t, router, "/api/does-not-exist")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) == indexHTML {
		t.Error("unknown API path returned the SPA fallback")
	}
}

func TestRouterWithoutStaticDirKeepsPreviousBehavior(t *testing.T) {
	router := NewRouter(newFixtureServer(), "")

	tests := map[string]int{
		"/topics":       http.StatusOK,
		"/api/topics":   http.StatusNotFound,
		"/unknown/page": http.StatusNotFound,
	}
	for path, want := range tests {
		if got := get(t, router, path).Code; got != want {
			t.Errorf("%s: expected status %d, got %d", path, want, got)
		}
	}
}
