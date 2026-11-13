package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestRequestDurationHistogramUsesSecondBuckets(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(RequestDurationView()),
	)
	prev := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	ResetInstruments()
	t.Cleanup(func() {
		_ = mp.Shutdown(context.Background())
		otel.SetMeterProvider(prev)
		ResetInstruments()
	})

	requestDuration.Record(context.Background(), 0.003)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}
			h, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("expected histogram data, got %T", m.Data)
			}
			for _, dp := range h.DataPoints {
				if dp.Count != 1 {
					continue
				}
				for i, count := range dp.BucketCounts {
					if count == 0 {
						continue
					}
					if i >= len(dp.Bounds) {
						t.Fatalf("3ms sample fell into overflow bucket %d with bounds %v", i, dp.Bounds)
					}
					if dp.Bounds[i] >= 0.01 {
						t.Fatalf("3ms sample landed in bucket with upper bound %v, want below 0.01 (bounds=%v)", dp.Bounds[i], dp.Bounds)
					}
					return
				}
			}
			t.Fatalf("no recorded data point found in histogram: %+v", h.DataPoints)
		}
	}
	t.Fatal("http.server.request.duration histogram not found")
}
