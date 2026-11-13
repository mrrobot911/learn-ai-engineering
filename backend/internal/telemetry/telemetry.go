package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// RequestDurationView configures second-based bucket boundaries for the HTTP
// server request duration histogram so Prometheus quantiles reflect seconds.
func RequestDurationView() sdkmetric.View {
	return sdkmetric.NewView(
		sdkmetric.Instrument{Name: "http.server.request.duration"},
		sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}},
	)
}

// Setup configures tracing, metrics and log export from the standard OTEL_*
// environment variables. When no OTLP endpoint is configured it is a no-op:
// providers stay at their defaults and the API behaves the same. The returned
// shutdown flushes and stops any configured providers.
func Setup(ctx context.Context) (shutdown func(context.Context) error, err error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") == "" {
		ConfigureLogger(nil)
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	var shutdowns []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			errs = append(errs, shutdowns[i](ctx))
		}
		return errors.Join(errs...)
	}

	traceExp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	shutdowns = append(shutdowns, tp.Shutdown)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	metricExp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		_ = shutdown(ctx)
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
		sdkmetric.WithExemplarFilter(exemplar.TraceBasedFilter),
		sdkmetric.WithView(RequestDurationView()),
	)
	shutdowns = append(shutdowns, mp.Shutdown)
	otel.SetMeterProvider(mp)

	logExp, err := otlploghttp.New(ctx)
	if err != nil {
		_ = shutdown(ctx)
		return nil, err
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	shutdowns = append(shutdowns, lp.Shutdown)
	global.SetLoggerProvider(lp)

	// Request logs go to stdout as JSON and are also exported over OTLP.
	ConfigureLogger(lp)

	_ = semconv.SchemaURL
	return shutdown, nil
}

// ConfigureLogger sets the default slog logger to emit JSON to stdout and,
// when a logger provider is configured, to export logs over OTLP via otelslog.
func ConfigureLogger(lp *sdklog.LoggerProvider) {
	stdout := slog.NewJSONHandler(LogWriter, nil)
	if lp == nil {
		logger := slog.New(stdout)
		slog.SetDefault(logger)
		SetRequestLogger(logger)
		return
	}
	otelHandler := otelslog.NewHandler("backend", otelslog.WithLoggerProvider(lp))
	logger := slog.New(newMultiHandler(stdout, otelHandler))
	slog.SetDefault(logger)
	SetRequestLogger(logger)
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, hh := range h.handlers {
		if hh.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, hh := range h.handlers {
		if hh.Enabled(ctx, r.Level) {
			errs = append(errs, hh.Handle(ctx, r.Clone()))
		}
	}
	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, 0, len(h.handlers))
	for _, hh := range h.handlers {
		hs = append(hs, hh.WithAttrs(attrs))
	}
	return &multiHandler{handlers: hs}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, 0, len(h.handlers))
	for _, hh := range h.handlers {
		hs = append(hs, hh.WithGroup(name))
	}
	return &multiHandler{handlers: hs}
}
