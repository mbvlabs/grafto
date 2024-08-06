package config

import "github.com/caarlos0/env/v10"

type Telemetry struct {
	TenantID string `env:"TENANT_ID"`
	SinkURL  string `env:"SINK_URL"`
}

func newTelemetry() Telemetry {
	telemetryCfg := Telemetry{}

	if err := env.Parse(&telemetryCfg); err != nil {
		panic(err)
	}

	return telemetryCfg
}
