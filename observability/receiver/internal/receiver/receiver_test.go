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
		switch r.URL.Path {
		case "/api/search":
			_ = json.NewEncoder(w).Encode(map[string]any{"traces": []any{map[string]any{"traceID": "expected-trace-id", "rootServiceName": "backend"}}})
		case "/api/traces/expected-trace-id":
			_ = json.NewEncoder(w).Encode(map[string]any{"batches": []any{map[string]any{"traceID": "expected-trace-id-full"}}})
		default:
			t.Fatalf("unexpected tempo path: %s", r.URL.Path)
		}
	}))
	defer tempo.Close()

	bundleDir := t.TempDir()
	repoDir := t.TempDir()
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	promptFile := filepath.Join(t.TempDir(), "prompt.txt")
	script := "#!/bin/sh\nprintf '%s\n' \"$@\" > " + argsFile + "\nprintf '%s' \"$2\" > " + promptFile + "\necho opencode-out\n"
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
	srv.Wait()

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
		t.Fatalf("traces.json missing search trace id: %s", string(tracesBytes))
	}
	if !strings.Contains(string(tracesBytes), "expected-trace-id-full") {
		t.Fatalf("traces.json missing fetched trace content: %s", string(tracesBytes))
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
	if _, err := os.Stat(filepath.Join(dir, "opencode.log")); err == nil {
		t.Fatal("raw opencode output must not be stored in bundle")
	}
	resultBytes, err := os.ReadFile(filepath.Join(dir, "agent-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(resultBytes), "output_tail") {
		t.Fatalf("agent-result.json missing tail: %s", string(resultBytes))
	}

	// Sentinel must be absent from bundle and generated prompt/args.
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

func TestRepeatedNotificationStartsAgentOnce(t *testing.T) {
	bundleDir := t.TempDir()
	repoDir := t.TempDir()
	countFile := filepath.Join(t.TempDir(), "count")
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	script := "#!/bin/sh\nif [ -f " + countFile + " ]; then c=$(cat " + countFile + "); else c=0; fi; c=$((c+1)); echo $c > " + countFile + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := New(Config{BundleRoot: bundleDir, RepoDir: repoDir, LokiURL: "http://127.0.0.1:1", TempoURL: "http://127.0.0.1:1", OpencodeBin: fakeBin, APIKey: "k"})
	srv.now = func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }
	srv.client = &http.Client{Timeout: 50 * time.Millisecond}
	payload := `{"status":"firing","alerts":[{"status":"firing","fingerprint":"fp1","labels":{"alertname":"QuizSubmission5xxRateHigh"},"startsAt":"2026-11-14T11:55:00Z"}]}`

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(payload))
		w := httptest.NewRecorder()
		srv.Routes().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("call %d: status=%d body=%s", i, w.Code, w.Body.String())
		}
	}
	srv.Wait()
	b, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("agent not started: %v", err)
	}
	if strings.TrimSpace(string(b)) != "1" {
		t.Fatalf("expected 1 agent run, got %q", string(b))
	}
}

func TestResolvedNotificationStoresBundleWithoutAgent(t *testing.T) {
	bundleDir := t.TempDir()
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	marker := filepath.Join(t.TempDir(), "called")
	script := "#!/bin/sh\necho called > " + marker + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := New(Config{BundleRoot: bundleDir, LokiURL: "http://127.0.0.1:1", TempoURL: "http://127.0.0.1:1", OpencodeBin: fakeBin, APIKey: "k"})
	srv.now = func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }
	srv.client = &http.Client{Timeout: 50 * time.Millisecond}

	payload := `{"status":"resolved","alerts":[{"status":"resolved","labels":{"alertname":"QuizSubmission5xxRateHigh"},"startsAt":"2026-11-14T11:55:00Z","endsAt":"2026-11-14T11:58:00Z"}]}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	srv.Wait()
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("opencode should not have been started for resolved alert")
	}
	entries, _ := os.ReadDir(bundleDir)
	if len(entries) != 1 {
		t.Fatalf("expected bundle stored, got %v", entries)
	}
	metaBytes, err := os.ReadFile(filepath.Join(bundleDir, entries[0].Name(), "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(metaBytes), "not_firing") {
		t.Fatalf("expected not_firing recorded, got: %s", string(metaBytes))
	}
}

func TestAgentFailureRecordedWithoutRawOutput(t *testing.T) {
	bundleDir := t.TempDir()
	repoDir := t.TempDir()
	sentinel := "SENTINEL_API_KEY_12345"
	fakeBin := filepath.Join(t.TempDir(), "fake-opencode")
	script := "#!/bin/sh\necho 'fail " + sentinel + "'\nexit 2\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := New(Config{BundleRoot: bundleDir, RepoDir: repoDir, LokiURL: "http://127.0.0.1:1", TempoURL: "http://127.0.0.1:1", OpencodeBin: fakeBin, APIKey: sentinel})
	srv.now = func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }
	srv.client = &http.Client{Timeout: 50 * time.Millisecond}

	payload := `{"status":"firing","alerts":[{"status":"firing","fingerprint":"fp-fail","labels":{"alertname":"QuizSubmission5xxRateHigh"},"startsAt":"2026-11-14T11:55:00Z"}]}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 once bundle stored, got %d body=%s", w.Code, w.Body.String())
	}
	srv.Wait()
	entries, _ := os.ReadDir(bundleDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 bundle, got %v", entries)
	}
	dir := filepath.Join(bundleDir, entries[0].Name())
	if _, err := os.Stat(filepath.Join(dir, "opencode.log")); err == nil {
		t.Fatal("raw opencode output must not be stored in bundle")
	}
	resultBytes, err := os.ReadFile(filepath.Join(dir, "agent-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(resultBytes), "error") {
		t.Fatalf("expected failure recorded, got: %s", string(resultBytes))
	}
	checkNoSentinel(t, dir, sentinel)
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
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	srv.Wait()
	entries, err := os.ReadDir(bundleDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 bundle dir, got %v err=%v", entries, err)
	}
	metaBytes, err := os.ReadFile(filepath.Join(bundleDir, entries[0].Name(), "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(metaBytes), "missing API key") {
		t.Fatalf("expected missing key recorded in metadata, got: %s", string(metaBytes))
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("opencode should not have been started")
	}
}
