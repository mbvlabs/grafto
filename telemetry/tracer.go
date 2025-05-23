package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// setupTraceProvider creates and configures a trace provider based on configuration
func setupTraceProvider(ctx context.Context, cfg Config, resource *resource.Resource) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if cfg.IsProduction() {
		// Production: Use OTLP HTTP exporter
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.OtlpEndpoint),
		}
		if cfg.OtlpInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	} else {
		// Development: Use stdout exporter
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	}

	// Configure trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio)),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// getTracer returns a tracer for the given name
func getTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
