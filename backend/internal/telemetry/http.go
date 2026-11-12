package telemetry

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/mrrobot911/learn-ai-engineering/internal/telemetry"

var (
	LogWriter io.Writer = os.Stdout

	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
	quizSubmissions metric.Int64Counter
)

func init() {
	initInstruments()
}

func initInstruments() {
	meter := otel.Meter(instrumentationName)
	requestCount, _ = meter.Int64Counter("http.server.request.count", metric.WithDescription("Number of HTTP requests"))
	requestDuration, _ = meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"), metric.WithDescription("Duration of HTTP requests"))
	quizSubmissions, _ = meter.Int64Counter("quiz.submissions", metric.WithDescription("Number of quiz submissions"))
}

// ResetInstruments rebinds instruments to the current global meter provider.
// Tests that install a meter provider after package init should call it.
func ResetInstruments() { initInstruments() }

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Middleware instruments HTTP requests with tracing, metrics and structured
// logging. The route template recorded is the mux pattern (r.Pattern) without
// the /api prefix, or the request path when no pattern is available.
func Middleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(LogWriter, nil))
	tracer := otel.Tracer(instrumentationName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := tracer.Start(r.Context(), "HTTP "+r.Method, trace.WithSpanKind(trace.SpanKindServer))

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		rwReq := r.WithContext(ctx)
		next.ServeHTTP(sw, rwReq)

		route := routeTemplate(rwReq)
		if route == "" {
			route = r.URL.Path
		}
		span.SetName(r.Method + " " + route)
		status := strconv.Itoa(sw.status)
		attrs := []attribute.KeyValue{
			attribute.String("method", r.Method),
			attribute.String("route", route),
			attribute.String("status", status),
		}
		span.SetAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", sw.status),
		)
		if sw.status >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		dur := time.Since(start).Seconds()
		attrSet := attribute.NewSet(attrs...)
		requestCount.Add(ctx, 1, metric.WithAttributeSet(attrSet))
		requestDuration.Record(ctx, dur, metric.WithAttributeSet(attrSet))

		if r.Method == http.MethodPost && route == "/topics/{topicSlug}/quiz/submissions" {
			quizSubmissions.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(
				attribute.String("route", route),
				attribute.String("status", status),
			)))
		}

		sc := trace.SpanContextFromContext(ctx)
		logger.InfoContext(ctx, "request",
			"method", r.Method,
			"route", route,
			"status", sw.status,
			"duration", time.Since(start).String(),
			"trace_id", sc.TraceID().String(),
			"span_id", sc.SpanID().String(),
		)
		span.End()
	})
}

func routeTemplate(r *http.Request) string {
	// Prefer pattern set by ServeMux on the served request.
	if p := r.Pattern; p != "" {
		// r.Pattern is "METHOD /path" in recent Go; strip method prefix.
		if i := len(r.Method) + 1; len(p) > i && p[:len(r.Method)] == r.Method && p[len(r.Method)] == ' ' {
			return p[i:]
		}
		// fallback: take everything after first space
		for i := 0; i < len(p); i++ {
			if p[i] == ' ' {
				return p[i+1:]
			}
		}
		return p
	}
	return ""
}
