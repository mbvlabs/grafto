package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Telemetry holds all telemetry providers and utilities
type Telemetry struct {
	tracer   trace.Tracer
	meter    metric.Meter
	logger   *Logger
	metrics  *ApplicationMetrics
	shutdown func(context.Context) error
}

// New creates and initializes a new Telemetry instance
func New(ctx context.Context, cfg Config) (*Telemetry, error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error

	// Setup tracing if enabled
	var tracer trace.Tracer
	if cfg.EnableTracing {
		tp, err := setupTraceProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup trace provider: %w", err)
		}

		tracer = getTracer(cfg.ServiceName)
		shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	}
	if !cfg.EnableTracing {
		tracer = tracenoop.NewTracerProvider().Tracer(cfg.ServiceName)
	}

	// Setup metrics if enabled
	var meter metric.Meter
	var appMetrics *ApplicationMetrics
	if cfg.EnableMetrics {
		mp, err := setupMeterProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup meter provider: %w", err)
		}

		meter = getMeter(cfg.ServiceName)

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
	if !cfg.EnableMetrics {
		noopProvider := noop.NewMeterProvider()
		meter = noopProvider.Meter(cfg.ServiceName)
	}

	// Setup enhanced logger
	logger := NewLogger(!cfg.IsProduction())
	slog.SetDefault(logger.Logger)

	// Create shutdown function that calls all provider shutdowns
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

// Tracer returns the OpenTelemetry tracer
func (t *Telemetry) Tracer() trace.Tracer {
	return t.tracer
}

// Meter returns the OpenTelemetry meter
func (t *Telemetry) Meter() metric.Meter {
	return t.meter
}

// Logger returns the enhanced logger with trace context support
func (t *Telemetry) Logger() *Logger {
	return t.logger
}

// Metrics returns the application-specific metrics
func (t *Telemetry) Metrics() *ApplicationMetrics {
	return t.metrics
}

// Shutdown gracefully shuts down all telemetry providers
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.shutdown != nil {
		return t.shutdown(ctx)
	}
	return nil
}

// StartSpan is a convenience method to start a new span
func (t *Telemetry) StartSpan(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, spanName, opts...)
	return ctx, span
}

// RecordMetric is a convenience method to record a metric value
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
