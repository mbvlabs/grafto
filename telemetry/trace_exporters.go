package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OtlpHttpTraceExporter struct {
	OtlpEndpoint string
	exporter     sdktrace.SpanExporter
}

func NewOtlpHttpTraceExporter() *OtlpHttpTraceExporter {
	return &OtlpHttpTraceExporter{}
}

func (o *OtlpHttpTraceExporter) Name() string {
	return "otlp-http"
}

func (o *OtlpHttpTraceExporter) GetSpanExporter(
	ctx context.Context,
	res *resource.Resource,
) (sdktrace.SpanExporter, error) {
	endpoint := strings.TrimPrefix(o.OtlpEndpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
		// otlptracehttp.WithInsecure(),
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create OTLP HTTP trace exporter: %w",
			err,
		)
	}

	o.exporter = exporter
	return exporter, nil
}

func (o *OtlpHttpTraceExporter) Shutdown(ctx context.Context) error {
	if o.exporter != nil {
		slog.InfoContext(ctx, "OTLP HTTP Trace Exporter shutting down...")
		return o.exporter.Shutdown(ctx)
	}
	return nil
}

var _ TraceExporter = new(OtlpHttpTraceExporter)

type noopSpanExporter struct{}

func (e *noopSpanExporter) ExportSpans(
	ctx context.Context,
	spans []sdktrace.ReadOnlySpan,
) error {
	return nil
}

func (e *noopSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

type NoopTraceExporter struct {
	exporter sdktrace.SpanExporter
}

func (n *NoopTraceExporter) Name() string {
	return "noop"
}

// GetSpanExporter returns an instance of noopSpanExporter.
func (n *NoopTraceExporter) GetSpanExporter(
	ctx context.Context,
	res *resource.Resource,
) (sdktrace.SpanExporter, error) {
	return &noopSpanExporter{}, nil
}

func (n *NoopTraceExporter) Shutdown(ctx context.Context) error {
	return nil
}

var _ TraceExporter = new(NoopTraceExporter)
