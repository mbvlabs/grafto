package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// setupMeterProvider creates and configures a meter provider based on configuration
func setupMeterProvider(ctx context.Context, cfg Config, resource *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var exporter sdkmetric.Exporter
	var err error

	if cfg.IsProduction() {
		// Production: Use OTLP HTTP exporter
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(cfg.OtlpEndpoint),
		}
		if cfg.OtlpInsecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		exporter, err = otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		}
	} else {
		// Development: Use stdout exporter
		exporter, err = stdoutmetric.New(
			stdoutmetric.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout metric exporter: %w", err)
		}
	}

	// Configure meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(cfg.GetMetricsPushInterval()),
		)),
	)

	// Set as global meter provider
	otel.SetMeterProvider(mp)

	return mp, nil
}

// getMeter returns a meter for the given name
func getMeter(name string) metric.Meter {
	return otel.Meter(name)
}

// ApplicationMetrics holds common application metrics
type ApplicationMetrics struct {
	HTTPRequestCount    metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	AuthAttempts        metric.Int64Counter
	RegistrationCount   metric.Int64Counter
	EmailsSent          metric.Int64Counter
	DBConnections       metric.Int64UpDownCounter
	BackgroundJobs      metric.Int64Counter
}

// NewApplicationMetrics creates and initializes application-specific metrics
func NewApplicationMetrics(meter metric.Meter) (*ApplicationMetrics, error) {
	httpRequestCount, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request counter: %w", err)
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request duration histogram: %w", err)
	}

	authAttempts, err := meter.Int64Counter(
		"auth_attempts_total",
		metric.WithDescription("Total number of authentication attempts"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth attempts counter: %w", err)
	}

	registrationCount, err := meter.Int64Counter(
		"registrations_total",
		metric.WithDescription("Total number of user registrations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration counter: %w", err)
	}

	emailsSent, err := meter.Int64Counter(
		"emails_sent_total",
		metric.WithDescription("Total number of emails sent"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create emails sent counter: %w", err)
	}

	dbConnections, err := meter.Int64UpDownCounter(
		"db_connections_active",
		metric.WithDescription("Number of active database connections"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB connections gauge: %w", err)
	}

	backgroundJobs, err := meter.Int64Counter(
		"background_jobs_total",
		metric.WithDescription("Total number of background jobs processed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create background jobs counter: %w", err)
	}

	return &ApplicationMetrics{
		HTTPRequestCount:    httpRequestCount,
		HTTPRequestDuration: httpRequestDuration,
		AuthAttempts:        authAttempts,
		RegistrationCount:   registrationCount,
		EmailsSent:          emailsSent,
		DBConnections:       dbConnections,
		BackgroundJobs:      backgroundJobs,
	}, nil
}
