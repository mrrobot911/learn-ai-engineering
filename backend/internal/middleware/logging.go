package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// LogWriter is the destination for request logs. Tests may replace it.
var LogWriter io.Writer = os.Stdout

// Logging logs requests as structured JSON with slog. It carries trace and span
// ids when a span is present in the request context. It does not log bodies,
// cookies, passwords, emails, or other sensitive values.
func Logging(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(LogWriter, nil))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		route := r.Pattern
		if route == "" {
			route = r.URL.Path
		} else {
			// Strip "METHOD " prefix from mux pattern.
			if i := len(r.Method) + 1; len(route) > i && route[:len(r.Method)] == r.Method && route[len(r.Method)] == ' ' {
				route = route[i:]
			}
		}
		sc := trace.SpanContextFromContext(r.Context())
		logger.InfoContext(r.Context(), "request",
			"method", r.Method,
			"route", route,
			"status", rw.statusCode,
			"duration", time.Since(start).String(),
			"trace_id", sc.TraceID().String(),
			"span_id", sc.SpanID().String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
