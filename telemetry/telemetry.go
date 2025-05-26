package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/sync/errgroup"
)

type Telemetry struct {
	AppTracerProvider *sdktrace.TracerProvider
	AppMetricProvider *sdkmetric.MeterProvider
	shutdown          []func(context.Context) error
}

func New(
	ctx context.Context,
	svcVersion string,
	appLogExporter LogExporter,
	appTraceExporter TraceExporter,
	appMetricExporter MetricExporter,
) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.Cfg.ServiceName),
			semconv.ServiceVersion(svcVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error

	logHandler, loghandlerShutdown := NewLogger(
		ctx,
		appLogExporter,
	)

	slog.SetDefault(logHandler)

	shutdownFuncs = append(shutdownFuncs, loghandlerShutdown)

	tp, err := NewTraceProvider(ctx, res, appTraceExporter, 1.0)
	if err != nil {
		return nil, fmt.Errorf("failed to setup trace provider: %w", err)
	}

	otel.SetTracerProvider(tp)

	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)

	mp, err := newMeterProvider(ctx, res, appMetricExporter, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to setup meter provider: %w", err)
	}

	otel.SetMeterProvider(mp)

	return &Telemetry{
		tp,
		mp,
		shutdownFuncs,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	ctxDeadline, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	eg := errgroup.Group{}

	for _, shutdown := range t.shutdown {
		eg.Go(func() error {
			return shutdown(ctxDeadline)
		})
	}

	return eg.Wait()
}
