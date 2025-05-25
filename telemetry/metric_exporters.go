package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
)

type OtlpHttpMetricExporter struct {
	OtlpEndpoint string
	exporter     sdkmetric.Exporter
}

func NewOtlpHttpMetricExporter() *OtlpHttpMetricExporter {
	return &OtlpHttpMetricExporter{}
}

func (o *OtlpHttpMetricExporter) Name() string {
	return "otlp-http"
}

func (o *OtlpHttpMetricExporter) GetSdkMetricExporter(
	ctx context.Context,
	res *resource.Resource,
) (sdkmetric.Exporter, error) {
	endpoint := strings.TrimPrefix(o.OtlpEndpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath(
			"/v1/metrics",
		),
		otlpmetrichttp.WithInsecure(),
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create OTLP HTTP metric exporter: %w",
			err,
		)
	}

	o.exporter = exporter
	return exporter, nil
}

func (o *OtlpHttpMetricExporter) Shutdown(ctx context.Context) error {
	if o.exporter != nil {
		slog.InfoContext(ctx, "OTLP HTTP Metric Exporter shutting down...")
		return o.exporter.Shutdown(ctx)
	}
	return nil
}

var _ MetricExporter = new(OtlpHttpMetricExporter)

type noopSdkMetricExporter struct{}

func (n *noopSdkMetricExporter) Aggregation(
	sdkmetric.InstrumentKind,
) sdkmetric.Aggregation {
	return nil
}

func (n *noopSdkMetricExporter) Export(
	context.Context,
	*metricdata.ResourceMetrics,
) error {
	return nil
}

func (n *noopSdkMetricExporter) ForceFlush(context.Context) error {
	return nil
}

func (n *noopSdkMetricExporter) Shutdown(context.Context) error {
	return nil
}

func (n *noopSdkMetricExporter) Temporality(
	sdkmetric.InstrumentKind,
) metricdata.Temporality {
	return 0
}

var _ sdkmetric.Exporter = new(noopSdkMetricExporter)

type NoopMetricExporter struct {
	exporter sdkmetric.Exporter
}

func (n *NoopMetricExporter) Name() string {
	return "noop"
}

func (n *NoopMetricExporter) GetSdkMetricExporter(
	ctx context.Context,
	res *resource.Resource,
) (sdkmetric.Exporter, error) {
	return &noopSdkMetricExporter{}, nil
}

func (n *NoopMetricExporter) Shutdown(ctx context.Context) error {
	return nil
}

var _ MetricExporter = new(NoopMetricExporter)

type BetterStackMetricExporter struct {
	Endpoint    string
	SourceToken string
	exporter    sdkmetric.Exporter
}

func NewBetterStackMetricExporter(
	endpoint, sourceToken string,
) *BetterStackMetricExporter {
	return &BetterStackMetricExporter{
		Endpoint:    endpoint,
		SourceToken: sourceToken,
	}
}

func (b *BetterStackMetricExporter) Name() string {
	return "betterstack"
}

func (b *BetterStackMetricExporter) GetSdkMetricExporter(
	ctx context.Context,
	res *resource.Resource,
) (sdkmetric.Exporter, error) {
	endpoint := strings.TrimPrefix(b.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath(
			"/metrics",
		),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + b.SourceToken,
		}),
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create BetterStack metric exporter: %w",
			err,
		)
	}

	b.exporter = exporter
	return exporter, nil
}

func (b *BetterStackMetricExporter) Shutdown(ctx context.Context) error {
	if b.exporter != nil {
		slog.InfoContext(ctx, "BetterStack Metric Exporter shutting down...")
		return b.exporter.Shutdown(ctx)
	}
	return nil
}

var _ MetricExporter = new(BetterStackMetricExporter)
