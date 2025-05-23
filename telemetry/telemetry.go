package telemetry

import (
	"context"
	"fmt"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type Telemetry struct {
	tracer   trace.Tracer
	meter    metric.Meter
	logger   *Logger
	metrics  *ApplicationMetrics
	shutdown func(context.Context) error
}

func New(ctx context.Context) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.Cfg.ServiceName),
			semconv.ServiceVersion(config.Cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error

	// Setup tracing if enabled
	var tracer trace.Tracer
	if config.Cfg.EnableTracing {
		tp, err := setupTraceProvider(ctx, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup trace provider: %w", err)
		}

		tracer = getTracer(config.Cfg.ServiceName)
		shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	}
	if !config.Cfg.EnableTracing {
		tracer = tracenoop.NewTracerProvider().Tracer(config.Cfg.ServiceName)
	}

	var meter metric.Meter
	var appMetrics *ApplicationMetrics
	if config.Cfg.EnableMetrics {
		mp, err := setupMeterProvider(ctx, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup meter provider: %w", err)
		}

		meter = getMeter(config.Cfg.ServiceName)

		// Create application-specific metrics
		appMetrics, err = NewApplicationMetrics(meter)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create application metrics: %w",
				err,
			)
		}

		shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	}
	if !config.Cfg.EnableMetrics {
		noopProvider := noop.NewMeterProvider()
		meter = noopProvider.Meter(config.Cfg.ServiceName)
	}

	// Setup enhanced logger
	logger := NewLogger(config.Cfg.Environment == config.DEV_ENVIRONMENT)

	shutdown := func(ctx context.Context) error {
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	}

	return &Telemetry{
		tracer:   tracer,
		meter:    meter,
		logger:   logger,
		metrics:  appMetrics,
		shutdown: shutdown,
	}, nil
}

func (t *Telemetry) Tracer() trace.Tracer {
	return t.tracer
}

func (t *Telemetry) Meter() metric.Meter {
	return t.meter
}

func (t *Telemetry) Logger() *Logger {
	return t.logger
}

func (t *Telemetry) Metrics() *ApplicationMetrics {
	return t.metrics
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.shutdown != nil {
		return t.shutdown(ctx)
	}
	return nil
}

func (t *Telemetry) StartSpan(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, spanName, opts...)
	return ctx, span
}

func (t *Telemetry) RecordMetric(
	ctx context.Context,
	name string,
	value int64,
	attrs ...metric.AddOption,
) {
	if counter, err := t.meter.Int64Counter(name); err == nil {
		counter.Add(ctx, value, attrs...)
	}
}
