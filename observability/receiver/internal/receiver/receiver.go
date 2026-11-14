package receiver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	BundleRoot   string
	RepoDir      string
	LokiURL      string
	TempoURL     string
	OpencodeBin  string
	Provider     string
	Model        string
	APIKey       string
	APIKeySource string
}

type Server struct {
	cfg    Config
	client *http.Client
	now    func() time.Time
	run    func(ctx context.Context, name string, args []string, dir string, env []string) ([]byte, error)
}

func New(cfg Config) *Server {
	if cfg.BundleRoot == "" {
		cfg.BundleRoot = "/data/incidents"
	}
	if cfg.RepoDir == "" {
		cfg.RepoDir = "/repo"
	}
	if cfg.LokiURL == "" {
		cfg.LokiURL = "http://loki:3100"
	}
	if cfg.TempoURL == "" {
		cfg.TempoURL = "http://tempo:3200"
	}
	if cfg.OpencodeBin == "" {
		cfg.OpencodeBin = "opencode"
	}
	return &Server{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
		now:    time.Now,
		run: func(ctx context.Context, name string, args []string, dir string, env []string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			cmd.Env = env
			return cmd.CombinedOutput()
		},
	}
}

type alertPayload struct {
	Status string  `json:"status"`
	Alerts []alert `json:"alerts"`
}

type alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/alerts", s.handleAlerts)
	return mux
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var payload alertPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(payload.Alerts) == 0 {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"ignored"}`))
		return
	}
	// Only act on firing alerts; store resolved notifications without starting agent.
	firing := payload.Status == "firing"
	if !firing {
		for _, a := range payload.Alerts {
			if a.Status == "firing" {
				firing = true
				break
			}
		}
	}
	bundlePath, err := s.process(r.Context(), body, payload, firing)
	if err != nil {
		log.Printf("incident handling failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "stored", "bundle": bundlePath})
}

func (s *Server) process(ctx context.Context, raw []byte, payload alertPayload, startAgent bool) (string, error) {
	start, end := incidentWindow(payload, s.now())
	id := incidentID(payload, s.now())
	dir := filepath.Join(s.cfg.BundleRoot, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create bundle dir: %w", err)
	}
	logs, logsErr := s.fetchLoki(ctx, start, end)
	traces, tracesErr := s.fetchTempo(ctx, start, end)

	meta := map[string]any{
		"incident_id": id,
		"bundle_dir":  dir,
		"volume_path": dir,
		"created_at":  s.now().UTC().Format(time.RFC3339),
		"window": map[string]string{
			"start": start.UTC().Format(time.RFC3339Nano),
			"end":   end.UTC().Format(time.RFC3339Nano),
		},
		"provider": s.cfg.Provider,
		"model":    s.cfg.Model,
	}
	if logsErr != nil {
		meta["logs_error"] = logsErr.Error()
	}
	if tracesErr != nil {
		meta["traces_error"] = tracesErr.Error()
	}
	if err := writeJSON(filepath.Join(dir, "alert.json"), json.RawMessage(raw)); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(dir, "metadata.json"), meta); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(dir, "logs.json"), logs); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(dir, "traces.json"), traces); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "bundle.path"), []byte(dir+"\n"), 0o644); err != nil {
		return "", err
	}

	if !startAgent {
		return dir, nil
	}
	if strings.TrimSpace(s.cfg.APIKey) == "" {
		return dir, errors.New("missing API key: set OPENCODE_API_KEY (or the provider's documented equivalent) before starting the on-call agent")
	}
	prompt := buildPrompt(payload, dir, meta)
	if err := os.WriteFile(filepath.Join(dir, "prompt.txt"), []byte(prompt), 0o600); err != nil {
		return "", err
	}
	env := buildEnv(s.cfg)
	out, err := s.run(ctx, s.cfg.OpencodeBin, []string{"run", prompt}, s.cfg.RepoDir, env)
	_ = os.WriteFile(filepath.Join(dir, "opencode.log"), out, 0o600)
	if err != nil {
		return dir, fmt.Errorf("opencode run failed: %w", err)
	}
	return dir, nil
}

func incidentWindow(p alertPayload, now time.Time) (time.Time, time.Time) {
	start := now.Add(-10 * time.Minute)
	end := now
	zero := time.Time{}
	for _, a := range p.Alerts {
		if !a.StartsAt.IsZero() && (start.Equal(now.Add(-10*time.Minute)) || a.StartsAt.Before(start)) {
			start = a.StartsAt.Add(-2 * time.Minute)
		}
		if !a.EndsAt.IsZero() && !a.EndsAt.Equal(zero) && a.EndsAt.After(time.Unix(0, 0)) && a.EndsAt.Year() > 2000 {
			if a.EndsAt.After(end) {
				end = a.EndsAt.Add(1 * time.Minute)
			}
		}
	}
	if end.Before(start) {
		end = start.Add(15 * time.Minute)
	}
	return start, end
}

var nonWord = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func incidentID(p alertPayload, now time.Time) string {
	name := "incident"
	if len(p.Alerts) > 0 {
		if n := p.Alerts[0].Labels["alertname"]; n != "" {
			name = n
		}
	}
	name = nonWord.ReplaceAllString(name, "-")
	return fmt.Sprintf("%s-%s", name, now.UTC().Format("20060102T150405Z"))
}

func (s *Server) fetchLoki(ctx context.Context, start, end time.Time) (any, error) {
	u, err := url.Parse(strings.TrimRight(s.cfg.LokiURL, "/") + "/loki/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", `{service_name=~".+"}`)
	q.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	q.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	q.Set("limit", "500")
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return json.RawMessage(b), fmt.Errorf("loki status %d", resp.StatusCode)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return json.RawMessage(b), nil
	}
	return v, nil
}

func (s *Server) fetchTempo(ctx context.Context, start, end time.Time) (any, error) {
	u, err := url.Parse(strings.TrimRight(s.cfg.TempoURL, "/") + "/api/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	q.Set("limit", "20")
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return json.RawMessage(b), fmt.Errorf("tempo status %d", resp.StatusCode)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return json.RawMessage(b), nil
	}
	return v, nil
}

func buildPrompt(p alertPayload, bundleDir string, meta map[string]any) string {
	var b strings.Builder
	b.WriteString("You are the on-call agent for a local observability drill. ")
	b.WriteString("Work in the repository checkout on a temporary branch. Never deploy. ")
	b.WriteString("Never print or commit secrets or personal data. ")
	b.WriteString("Diagnose the incident using the evidence bundle and leave your diagnosis and fix as a commit or pull request on that branch.\n\n")
	b.WriteString("Incident bundle path: ")
	b.WriteString(bundleDir)
	b.WriteString("\n")
	if m, err := json.MarshalIndent(meta, "", "  "); err == nil {
		b.WriteString("Bundle metadata:\n")
		b.Write(m)
		b.WriteString("\n")
	}
	b.WriteString("Alert:\n")
	if j, err := json.MarshalIndent(p, "", "  "); err == nil {
		b.Write(j)
		b.WriteString("\n")
	}
	b.WriteString("\nRead alert.json, logs.json, traces.json and metadata.json in the bundle for full evidence.\n")
	return b.String()
}

func buildEnv(cfg Config) []string {
	env := os.Environ()
	// Do not add the API key value again beyond what is already in process env; it must come from env vars only.
	if cfg.Provider != "" {
		env = append(env, "OPENCODE_PROVIDER="+cfg.Provider)
	}
	if cfg.Model != "" {
		env = append(env, "OPENCODE_MODEL="+cfg.Model)
	}
	return env
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func LoadConfigFromEnv() Config {
	provider := strings.TrimSpace(os.Getenv("OPENCODE_PROVIDER"))
	model := strings.TrimSpace(os.Getenv("OPENCODE_MODEL"))
	key, src := resolveAPIKey()
	return Config{
		BundleRoot:   envOr("INCIDENT_BUNDLE_DIR", envOr("BUNDLE_DIR", "/data/incidents")),
		RepoDir:      envOr("REPO_DIR", "/repo"),
		LokiURL:      envOr("LOKI_URL", "http://loki:3100"),
		TempoURL:     envOr("TEMPO_URL", "http://tempo:3200"),
		OpencodeBin:  envOr("OPENCODE_BIN", "opencode"),
		Provider:     provider,
		Model:        model,
		APIKey:       key,
		APIKeySource: src,
	}
}

func resolveAPIKey() (string, string) {
	for _, name := range []string{"OPENCODE_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v, name
		}
	}
	return "", ""
}

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Routes())
}
