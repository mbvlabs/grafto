package telemetry

import (
	"context"

	"github.com/mbv-labs/grafto/config"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	tracer trace.Tracer
	name   string
}

type Otel struct {
	cfg           config.Config
	traceProvider *tracesdk.TracerProvider
}

func NewOtel(cfg config.Config) Otel {
	sampler := tracesdk.WithSampler(tracesdk.NeverSample())
	if cfg.App.Environment == config.PROD_ENVIRONMENT {
		sampler = tracesdk.WithSampler(tracesdk.AlwaysSample())
	}

	tp := tracesdk.NewTracerProvider(
		sampler,
	)

	return Otel{
		cfg,
		tp,
	}
}

func (o Otel) NewTracer(name string) Tracer {
	var t trace.Tracer
	if o.cfg.App.Environment == config.PROD_ENVIRONMENT {
		t = o.traceProvider.Tracer(name)
	}
	if o.cfg.App.Environment == config.DEV_ENVIRONMENT {
		t = NoopTracer{}
	}

	return Tracer{
		t,
		name,
	}
}

func (o Otel) Shutdown() error {
	return o.traceProvider.Shutdown(context.Background())
}

func (t Tracer) CreateSpan(
	ctx context.Context,
	name string,
) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}

func (t Tracer) CreateChildSpan(
	ctx context.Context,
	span trace.Span,
	name string,
) (context.Context, trace.Span) {
	return span.TracerProvider().Tracer(t.name).Start(ctx, name)
}
