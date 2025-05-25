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
