package receiver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIncidentBundleAndOpencodeCalledWithoutKeyLeak(t *testing.T) {
	loki := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/loki/api/v1/query_range") {
			t.Fatalf("unexpected loki path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   map[string]any{"resultType": "streams", "result": []any{map[string]any{"stream": map[string]string{"job": "backend"}, "values": []any{[]any{"1", "expected-log-line"}}}}},
		})
	}))
	defer loki.Close()

	tempo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Fatalf("unexpected tempo path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"traces": []any{map[string]any{"traceID": "expected-trace-id", "rootServiceName": "backend"}}})
	}))
	defer tempo.Close()

	bundleDir := t.TempDir()
	repoDir := t.TempDir()
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	promptFile := filepath.Join(t.TempDir(), "prompt.txt")
	script := "#!/bin/sh\nprintf '%s\n' \"$@\" > " + argsFile + "\nprintf '%s' \"$2\" > " + promptFile + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	sentinel := "SENTINEL_API_KEY_12345"
	cfg := Config{
		BundleRoot:  bundleDir,
		RepoDir:     repoDir,
		LokiURL:     loki.URL,
		TempoURL:    tempo.URL,
		OpencodeBin: fakeBin,
		Provider:    "test-provider",
		Model:       "test-model",
		APIKey:      sentinel,
	}
	srv := New(cfg)
	srv.now = func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }

	payload := map[string]any{
		"status": "firing",
		"alerts": []map[string]any{{
			"status":      "firing",
			"labels":      map[string]string{"alertname": "QuizSubmission5xxRateHigh", "severity": "warning"},
			"annotations": map[string]string{"summary": "Visitors cannot submit a quiz"},
			"startsAt":    "2026-11-14T11:55:00Z",
			"endsAt":      "0001-01-01T00:00:00Z",
		}},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	entries, err := os.ReadDir(bundleDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 bundle dir, got %v err=%v", entries, err)
	}
	dir := filepath.Join(bundleDir, entries[0].Name())

	logsBytes, err := os.ReadFile(filepath.Join(dir, "logs.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logsBytes), "expected-log-line") {
		t.Fatalf("logs.json missing expected content: %s", string(logsBytes))
	}
	tracesBytes, err := os.ReadFile(filepath.Join(dir, "traces.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tracesBytes), "expected-trace-id") {
		t.Fatalf("traces.json missing expected content: %s", string(tracesBytes))
	}

	argsBytes, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("fake opencode not called: %v", err)
	}
	if !strings.Contains(string(argsBytes), "run") {
		t.Fatalf("expected opencode run args, got: %s", string(argsBytes))
	}
	promptBytes, err := os.ReadFile(promptFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(promptBytes), dir) {
		t.Fatalf("prompt missing bundle path: %s", string(promptBytes))
	}
	if !strings.Contains(string(promptBytes), "QuizSubmission5xxRateHigh") {
		t.Fatalf("prompt missing incident context: %s", string(promptBytes))
	}

	// Sentinel must be absent from bundle, logs (receiver log output none) and generated prompt.
	checkNoSentinel(t, dir, sentinel)
	if strings.Contains(string(promptBytes), sentinel) {
		t.Fatal("sentinel API key leaked into prompt")
	}
	if strings.Contains(string(argsBytes), sentinel) {
		t.Fatal("sentinel API key leaked into args")
	}
	metaBytes, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if !strings.Contains(string(metaBytes), "test-provider") || !strings.Contains(string(metaBytes), "test-model") {
		t.Fatalf("metadata should record provider/model names: %s", string(metaBytes))
	}
}

func checkNoSentinel(t *testing.T, dir, sentinel string) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), sentinel) {
			t.Fatalf("sentinel API key leaked into %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMissingKeyFailsFastAndStartsNoAgent(t *testing.T) {
	bundleDir := t.TempDir()
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	marker := filepath.Join(t.TempDir(), "called")
	script := "#!/bin/sh\necho called > " + marker + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := New(Config{BundleRoot: bundleDir, LokiURL: "http://127.0.0.1:1", TempoURL: "http://127.0.0.1:1", OpencodeBin: fakeBin})
	srv.now = func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }
	srv.client = &http.Client{Timeout: 50 * time.Millisecond}

	payload := `{"status":"firing","alerts":[{"status":"firing","labels":{"alertname":"QuizSubmission5xxRateHigh"},"annotations":{"summary":"Visitors cannot submit a quiz"},"startsAt":"2026-11-14T11:55:00Z"}]}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "missing API key") {
		t.Fatalf("expected clear missing key message, got: %s", w.Body.String())
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("opencode should not have been started")
	}
}
