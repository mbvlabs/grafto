package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func setupMeterProvider(
	ctx context.Context,
	resource *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	var exporter sdkmetric.Exporter
	var err error

	// Use OTLP exporter for both prod and dev
	endpoint := config.Cfg.OtlpEndpoint
	// Remove protocol if present (OTLP HTTP exporter expects just host:port)
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath("/v1/metrics"),
	}
	if config.Cfg.OtlpInsecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err = otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create OTLP metric exporter: %w",
			err,
		)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(
				time.Duration(config.Cfg.MetricsPushInterval)*time.Second,
			),
		)),
	)

	otel.SetMeterProvider(mp)

	return mp, nil
}

func getMeter(name string) metric.Meter {
	return otel.Meter(name)
}

type ApplicationMetrics struct {
	HTTPRequestCount    metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	AuthAttempts        metric.Int64Counter
	RegistrationCount   metric.Int64Counter
	EmailsSent          metric.Int64Counter
	DBConnections       metric.Int64UpDownCounter
	BackgroundJobs      metric.Int64Counter
}

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
		metric.WithExplicitBucketBoundaries(
			0.0001,  // 0.1ms
			0.0005,  // 0.5ms
			0.001,   // 1ms
			0.005,   // 5ms
			0.01,    // 10ms
			0.025,   // 25ms
			0.05,    // 50ms
			0.1,     // 100ms
			0.25,    // 250ms
			0.5,     // 500ms
			1.0,     // 1s
			2.5,     // 2.5s
			5.0,     // 5s
			10.0,    // 10s
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create HTTP request duration histogram: %w",
			err,
		)
	}

	authAttempts, err := meter.Int64Counter(
		"auth_attempts_total",
		metric.WithDescription("Total number of authentication attempts"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create auth attempts counter: %w",
			err,
		)
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
		return nil, fmt.Errorf(
			"failed to create background jobs counter: %w",
			err,
		)
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
