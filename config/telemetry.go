package config

import "github.com/caarlos0/env/v10"

type Telemetry struct {
	TenantID string `env:"TENANT_ID"`
	SinkURL  string `env:"SINK_URL"`
}

func newTelemetry() Telemetry {
	telemetryCfg := Telemetry{}

	if err := env.ParseWithOptions(&telemetryCfg, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}

	return telemetryCfg
}
