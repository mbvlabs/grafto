package telemetry

import (
	"context"
	"fmt"
	"runtime"
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
			"failed to create OTLP metric exporter: %w",
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

func HTTPRequestsTotal() (metric.Int64Counter, error) {
	return GetMeter().Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
}

func HTTPRequestsInFlight() (metric.Int64UpDownCounter, error) {
	return GetMeter().Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Current number of HTTP requests being served"),
	)
}

func HTTPRequestDuration() (metric.Float64Histogram, error) {
	return GetMeter().Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
}

func GoGoroutinesMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"go_goroutines",
		metric.WithDescription("Number of goroutines that currently exist"),
		metric.WithUnit("1"),
	)
}

func GoThreadsMetric() (metric.Int64Gauge, error) {
	return GetMeter().Int64Gauge(
		"go_threads",
		metric.WithDescription("Number of OS threads created"),
		metric.WithUnit("1"),
	)
}

func GoMemstatsAllocBytesMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"go_memstats_alloc_bytes",
		metric.WithDescription("Number of bytes allocated and still in use"),
		metric.WithUnit("By"),
	)
}

func GoMemstatsHeapObjectsMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"go_memstats_heap_objects",
		metric.WithDescription("Number of allocated objects"),
		metric.WithUnit("1"),
	)
}

func GoMemstatsSysBytesMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"go_memstats_sys_bytes",
		metric.WithDescription("Number of bytes obtained from system"),
		metric.WithUnit("By"),
	)
}

func GoGCDurationMetric() (metric.Float64Histogram, error) {
	return GetMeter().Float64Histogram(
		"go_gc_duration_seconds",
		metric.WithDescription("A summary of the pause duration of garbage collection cycles"),
		metric.WithUnit("s"),
	)
}

func ProcessCPUSecondsTotalMetric() (metric.Float64Counter, error) {
	return GetMeter().Float64Counter(
		"process_cpu_seconds_total",
		metric.WithDescription("Total user and system CPU time spent in seconds"),
		metric.WithUnit("s"),
	)
}

func ProcessResidentMemoryBytesMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"process_resident_memory_bytes",
		metric.WithDescription("Resident memory size in bytes"),
		metric.WithUnit("By"),
	)
}

func ProcessVirtualMemoryBytesMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"process_virtual_memory_bytes",
		metric.WithDescription("Virtual memory size in bytes"),
		metric.WithUnit("By"),
	)
}

func ProcessOpenFDsMetric() (metric.Int64ObservableGauge, error) {
	return GetMeter().Int64ObservableGauge(
		"process_open_fds",
		metric.WithDescription("Number of open file descriptors"),
		metric.WithUnit("1"),
	)
}

func SetupRuntimeMetricsInCallback(meter metric.Meter) error {
	_, err := meter.Int64ObservableGauge(
		"go_goroutines",
		metric.WithDescription("Number of goroutines that currently exist"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				o.Observe(int64(runtime.NumGoroutine()))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_alloc_bytes",
		metric.WithDescription("Number of bytes allocated and still in use"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.Alloc))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_alloc_bytes",
		metric.WithDescription(
			"Number of heap bytes allocated and still in use",
		),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapAlloc))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_sys_bytes",
		metric.WithDescription("Number of heap bytes obtained from system"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapSys))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_idle_bytes",
		metric.WithDescription("Number of heap bytes waiting to be used"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapIdle))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_inuse_bytes",
		metric.WithDescription("Number of heap bytes that are in use"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapInuse))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_released_bytes",
		metric.WithDescription("Number of heap bytes released to OS"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapReleased))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_sys_bytes",
		metric.WithDescription("Number of bytes obtained from system"),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.Sys))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableCounter(
		"go_memstats_mallocs_total",
		metric.WithDescription("Total number of mallocs"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.Mallocs))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableCounter(
		"go_memstats_frees_total",
		metric.WithDescription("Total number of frees"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.Frees))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_heap_objects",
		metric.WithDescription("Number of allocated objects"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.HeapObjects))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_next_gc_bytes",
		metric.WithDescription(
			"Number of heap bytes when next garbage collection will take place",
		),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.NextGC))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Float64ObservableGauge(
		"go_memstats_last_gc_time_seconds",
		metric.WithDescription(
			"Number of seconds since 1970 of last garbage collection",
		),
		metric.WithUnit("s"),
		metric.WithFloat64Callback(
			func(ctx context.Context, o metric.Float64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				o.Observe(
					float64(m.LastGC) / 1e9,
				) // Convert nanoseconds to seconds
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_memstats_gc_sys_bytes",
		metric.WithDescription(
			"Number of bytes used for garbage collection system metadata",
		),
		metric.WithUnit("bytes"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//nolint:gosec // TODO
				o.Observe(int64(m.GCSys))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_threads",
		metric.WithDescription("Number of OS threads created"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				o.Observe(int64(runtime.GOMAXPROCS(0)))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	// GC duration requires special handling since we need to track pause times
	var previousNumGC uint32
	var totalPauseNs uint64

	_, err = meter.Float64ObservableGauge(
		"go_gc_duration_seconds_sum",
		metric.WithDescription(
			"Total pause duration of garbage collection cycles",
		),
		metric.WithUnit("s"),
		metric.WithFloat64Callback(
			func(ctx context.Context, o metric.Float64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				// Calculate cumulative pause time
				if m.NumGC > previousNumGC {
					// Add new pause times
					for i := previousNumGC; i < m.NumGC; i++ {
						idx := i % uint32(len(m.PauseNs))
						totalPauseNs += m.PauseNs[idx]
					}
					previousNumGC = m.NumGC
				}

				o.Observe(
					float64(totalPauseNs) / 1e9,
				) // Convert nanoseconds to seconds
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"go_gc_duration_seconds_count",
		metric.WithDescription("Number of garbage collection cycles"),
		metric.WithInt64Callback(
			func(ctx context.Context, o metric.Int64Observer) error {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				o.Observe(int64(m.NumGC))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	startTime := time.Now()
	_, err = meter.Float64ObservableGauge(
		"process_start_time_seconds",
		metric.WithDescription(
			"Start time of the process since unix epoch in seconds",
		),
		metric.WithUnit("s"),
		metric.WithFloat64Callback(
			func(ctx context.Context, o metric.Float64Observer) error {
				o.Observe(float64(startTime.Unix()))
				return nil
			},
		),
	)
	if err != nil {
		return err
	}

	return nil
}
