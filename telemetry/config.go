package telemetry

import "time"

// Config holds telemetry configuration options
type Config struct {
	EnableTracing       bool    `env:"TELEMETRY_ENABLE_TRACING" envDefault:"true"`
	EnableMetrics       bool    `env:"TELEMETRY_ENABLE_METRICS" envDefault:"true"`
	ServiceName         string  `env:"TELEMETRY_SERVICE_NAME" envDefault:"grafto"`
	ServiceVersion      string  `env:"TELEMETRY_SERVICE_VERSION" envDefault:"dev"`
	OtlpEndpoint        string  `env:"TELEMETRY_OTLP_ENDPOINT" envDefault:""`
	OtlpInsecure        bool    `env:"TELEMETRY_OTLP_INSECURE" envDefault:"true"`
	TraceSampleRatio    float64 `env:"TELEMETRY_TRACE_SAMPLE_RATIO" envDefault:"1.0"`
	MetricsPushInterval int     `env:"TELEMETRY_METRICS_PUSH_INTERVAL" envDefault:"30"`
}

// IsProduction returns true if we should use production exporters (OTLP)
func (c Config) IsProduction() bool {
	return c.OtlpEndpoint != ""
}

// GetMetricsPushInterval returns the metrics push interval as a duration
func (c Config) GetMetricsPushInterval() time.Duration {
	return time.Duration(c.MetricsPushInterval) * time.Second
}
