package telemetry

import (
	"context"
	"fmt"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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

	if config.Cfg.Environment == config.PROD_ENVIRONMENT {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(config.Cfg.OtlpEndpoint),
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
	}
	if config.Cfg.Environment == config.DEV_ENVIRONMENT {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create stdout trace exporter: %w",
				err,
			)
		}
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
