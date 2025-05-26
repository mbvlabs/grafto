package telemetry

import (
	"context"
	"fmt"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type TraceExporter interface {
	Name() string

	GetSpanExporter(
		ctx context.Context,
		res *resource.Resource,
	) (sdktrace.SpanExporter, error)

	Shutdown(ctx context.Context) error
}

func NewTraceProvider(
	ctx context.Context,
	resource *resource.Resource,
	traceExporter TraceExporter,
	sampleRatio float64,
) (*sdktrace.TracerProvider, error) {
	exporter, err := traceExporter.GetSpanExporter(ctx, resource)
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
			sdktrace.TraceIDRatioBased(sampleRatio),
		),
	)

	return tp, nil
}

func GetTracer() trace.Tracer {
	return otel.Tracer(config.Cfg.ServiceName)
}
