package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type telemetryFixture struct {
	exporter *tracetest.InMemoryExporter
	reader   *sdkmetric.ManualReader
	logs     *bytes.Buffer
}

func setupTelemetry(t *testing.T) *telemetryFixture {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevTP := otel.GetTracerProvider()
	prevMP := otel.GetMeterProvider()
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	telemetry.ResetInstruments()

	var logs bytes.Buffer
	prevWriter := telemetry.LogWriter
	telemetry.LogWriter = &logs
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		_ = mp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		telemetry.LogWriter = prevWriter
	})
	return &telemetryFixture{exporter: exporter, reader: reader, logs: &logs}
}

func testQuizServer(t *testing.T) http.Handler {
	t.Helper()
	topics := []repo.Topic{{ID: "topic-1", Slug: "intro", Title: "Intro", Position: 1, Status: repo.StatusPublished}}
	questions := []repo.Question{{
		ID: "q1", TopicID: "topic-1", Text: "Question?", Explanation: "Because.", Status: repo.StatusPublished,
		Options: []repo.Option{{ID: "o1", Text: "A", Position: 1, IsCorrect: true}, {ID: "o2", Text: "B", Position: 2}},
	}}
	contentRepo := repo.NewMemoryContentRepository(topics, nil, questions)
	attemptRepo := repo.NewMemoryAttemptRepository()
	svc := service.NewContentService(contentRepo, attemptRepo)
	h := handler.NewContentHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /topics/{topicSlug}/quiz", h.StartQuiz)
	mux.HandleFunc("POST /topics/{topicSlug}/quiz/submissions", h.SubmitQuiz)
	return telemetry.Middleware(handler.NewRouter(mux, t.TempDir()))
}

func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	return rm
}

func attrValue(set attribute.Set, key string) (string, bool) {
	for _, kv := range set.ToSlice() {
		if string(kv.Key) == key {
			return kv.Value.Emit(), true
		}
	}
	return "", false
}

func metricsDump(rm metricdata.ResourceMetrics) string {
	var b strings.Builder
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			fmt.Fprintf(&b, "metric:%s\n", m.Name)
			switch d := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, dp := range d.DataPoints {
					fmt.Fprintf(&b, " attrs=%v value=%d exemplars=%v\n", dp.Attributes.ToSlice(), dp.Value, dp.Exemplars)
				}
			case metricdata.Histogram[float64]:
				for _, dp := range d.DataPoints {
					fmt.Fprintf(&b, " attrs=%v count=%d sum=%v exemplars=%v\n", dp.Attributes.ToSlice(), dp.Count, dp.Sum, dp.Exemplars)
				}
			}
		}
	}
	return b.String()
}

func spansDump(spans tracetest.SpanStubs) string {
	var b strings.Builder
	for _, s := range spans {
		fmt.Fprintf(&b, "span:%s attrs=%v events=%v\n", s.Name, s.Attributes, s.Events)
	}
	return b.String()
}

func metricsContainRoute(rm metricdata.ResourceMetrics, route string) bool {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch d := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, dp := range d.DataPoints {
					if v, ok := attrValue(dp.Attributes, "route"); ok && v == route {
						return true
					}
				}
			case metricdata.Histogram[float64]:
				for _, dp := range d.DataPoints {
					if v, ok := attrValue(dp.Attributes, "route"); ok && v == route {
						return true
					}
				}
			}
		}
	}
	return false
}

func TestRouteTemplateUsesMuxPatternWithoutAPIPrefix(t *testing.T) {
	fx := setupTelemetry(t)
	srv := testQuizServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/topics/intro/quiz/submissions", strings.NewReader(`{"answers":[{"question_id":"q1","option_id":"o1"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	wantRoute := "/topics/{topicSlug}/quiz/submissions"
	spans := fx.exporter.GetSpans()
	found := false
	for _, s := range spans {
		for _, a := range s.Attributes {
			if string(a.Key) == "http.route" {
				if a.Value.AsString() == wantRoute {
					found = true
				}
				if strings.Contains(a.Value.AsString(), "/api") {
					t.Errorf("route leaked /api prefix: %v", a.Value.AsString())
				}
			}
		}
	}
	if !found {
		t.Fatalf("no span with route %q: %s", wantRoute, spansDump(spans))
	}

	logs := fx.logs.String()
	if !strings.Contains(logs, wantRoute) {
		t.Fatalf("logs missing route %q: %s", wantRoute, logs)
	}
	if strings.Contains(logs, "/api/topics/intro/quiz/submissions") {
		t.Fatalf("logs leaked /api path: %s", logs)
	}

	rm := collectMetrics(t, fx.reader)
	if !metricsContainRoute(rm, wantRoute) {
		t.Fatalf("metrics missing route %q: %s", wantRoute, metricsDump(rm))
	}
}

func TestQuizTraceAndLogsShareTraceID(t *testing.T) {
	fx := setupTelemetry(t)
	srv := testQuizServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/topics/intro/quiz/submissions", strings.NewReader(`{"answers":[{"question_id":"q1","option_id":"o1"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	spans := fx.exporter.GetSpans()
	var serverTraceID string
	for _, s := range spans {
		if strings.Contains(s.Name, "POST /topics/{topicSlug}/quiz/submissions") {
			serverTraceID = s.SpanContext.TraceID().String()
			break
		}
	}
	if serverTraceID == "" || serverTraceID == "00000000000000000000000000000000" {
		t.Fatalf("missing server span trace id: %s", spansDump(spans))
	}

	logs := fx.logs.String()
	if !strings.Contains(logs, serverTraceID) {
		t.Fatalf("logs do not contain trace id %q: %s", serverTraceID, logs)
	}
}

func TestTelemetryContainsNoSensitiveData(t *testing.T) {
	fx := setupTelemetry(t)
	srv := testQuizServer(t)

	email := "unique-user-9f2@example.test"
	password := "unique-password-9f2"
	cookie := "session=unique-cookie-9f2"
	bodyMarker := "unique-body-marker-9f2"
	body := fmt.Sprintf(`{"answers":[{"question_id":"q1","option_id":"o1"}],"marker":"%s","email":"%s","password":"%s"}`, bodyMarker, email, password)

	req := httptest.NewRequest(http.MethodPost, "/api/topics/intro/quiz/submissions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-CSRF-Token", "unique-csrf-9f2")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	spans := spansDump(fx.exporter.GetSpans())
	logs := fx.logs.String()
	metrics := metricsDump(collectMetrics(t, fx.reader))
	all := spans + "\n" + logs + "\n" + metrics
	for _, secret := range []string{email, password, cookie, bodyMarker, "unique-csrf-9f2"} {
		if strings.Contains(all, secret) {
			t.Errorf("telemetry leaked sensitive value %q in:\n%s", secret, all)
		}
	}
}

func TestQuizSubmissionHasChildSpans(t *testing.T) {
	fx := setupTelemetry(t)
	srv := testQuizServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/topics/intro/quiz/submissions", strings.NewReader(`{"answers":[{"question_id":"q1","option_id":"o1"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	spans := fx.exporter.GetSpans()
	names := map[string]bool{}
	for _, s := range spans {
		names[s.Name] = true
	}
	for _, want := range []string{"quiz.load_questions", "quiz.grade"} {
		if !names[want] {
			t.Errorf("missing child span %q: %s", want, spansDump(spans))
		}
	}
}

var _ = json.Marshal
