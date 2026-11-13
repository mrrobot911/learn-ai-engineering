package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/telemetry"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRequestLogsExportedOverOTLP(t *testing.T) {
	var logRequests atomic.Int32
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/logs") {
			logRequests.Add(1)
			_, _ = io.ReadAll(r.Body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithExemplarFilter(exemplar.TraceBasedFilter),
	)
	prevTP := otel.GetTracerProvider()
	prevMP := otel.GetMeterProvider()
	prevWriter := telemetry.LogWriter
	prevLogger := telemetry.RequestLogger()
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		_ = mp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		telemetry.LogWriter = prevWriter
		telemetry.SetRequestLogger(prevLogger)
	})
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	telemetry.ResetInstruments()

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("telemetry setup: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	srv := testQuizServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/topics/intro/quiz/submissions", strings.NewReader(`{"answers":[{"question_id":"q1","option_id":"o1"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown telemetry: %v", err)
	}
	if logRequests.Load() == 0 {
		t.Fatal("expected OTLP log export request to collector")
	}
}
