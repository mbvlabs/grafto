package config

import "github.com/caarlos0/env/v10"

type Telemetry struct {
	ServiceName  string `env:"TELEMETRY_SERVICE_NAME"`
	OtlpEndpoint string `env:"TELEMETRY_OTLP_ENDPOINT"`
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
