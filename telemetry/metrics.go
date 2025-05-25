package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/mbvlabs/grafto/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type MetricExporter interface {
	Name() string
	GetSdkMetricExporter(
		ctx context.Context,
		res *resource.Resource,
	) (sdkmetric.Exporter, error)
	Shutdown(ctx context.Context) error
}

func newMeterProvider(
	ctx context.Context,
	resource *resource.Resource,
	metricExporter MetricExporter,
	pushInterval time.Duration,
) (*sdkmetric.MeterProvider, error) {
	exporter, err := metricExporter.GetSdkMetricExporter(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create OTLP trace exporter: %w",
			err,
		)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(
				pushInterval,
			),
		)),
	)

	return mp, nil
}

func GetMeter() metric.Meter {
	return otel.Meter(config.Cfg.ServiceName)
}

func HttpRequestLatencyMetric() (metric.Float64Histogram, error) {
	reqLatMetric, err := GetMeter().Float64Histogram(

		"http_request_latency",
		metric.WithUnit("s"),
		// metric.WithExplicitBucketBoundaries(0.1, 0.5, 1.0),
		// metric.WithDescription("CUUUUSTOOOM"),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create HTTP request duration histogram: %w",
			err,
		)
	}

	return reqLatMetric, nil
}

func HttpRequestCountMetric() (metric.Int64Counter, error) {
	reqCountMetric, err := GetMeter().Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create HTTP request duration histogram: %w",
			err,
		)
	}

	return reqCountMetric, nil
}

// 	dbConnections, err := meter.Int64UpDownCounter(
// 		"db_connections_active",
// 		metric.WithDescription("Number of active database connections"),
// 		metric.WithUnit("1"),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create DB connections gauge: %w", err)
// 	}
//
// 	backgroundJobs, err := meter.Int64Counter(
// 		"background_jobs_total",
// 		metric.WithDescription("Total number of background jobs processed"),
// 		metric.WithUnit("1"),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"failed to create background jobs counter: %w",
// 			err,
// 		)
// 	}
