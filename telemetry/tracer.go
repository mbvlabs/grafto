package telemetry

import (
	"context"
	"fmt"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func setupTraceProvider(
	ctx context.Context,
	resource *resource.Resource,
) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	// Use OTLP exporter for both prod and dev
	endpoint := config.Cfg.OtlpEndpoint
	// // Remove protocol if present (OTLP HTTP exporter expects just host:port)
	// if strings.HasPrefix(endpoint, "http://") {
	// 	endpoint = strings.TrimPrefix(endpoint, "http://")
	// }
	// if strings.HasPrefix(endpoint, "https://") {
	// 	endpoint = strings.TrimPrefix(endpoint, "https://")
	// }

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
	}
	if config.Cfg.OtlpInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err = otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create OTLP trace exporter: %w",
			err,
		)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(
			sdktrace.TraceIDRatioBased(config.Cfg.TraceSampleRatio),
		),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func getTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
