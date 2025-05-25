package config

import (
	"os"

	"github.com/caarlos0/env/v10"
)

// Cfg instantiate a new cfg but can panic
var Cfg Config = NewConfig()

type Config struct {
	Database
	Authentication
	App
	Telemetry
	AwsAccessKeyID     string
	AwsSecretAccessKey string
}

type Telemetry struct {
	// EnableTracing bool   `env:"TELEMETRY_ENABLE_TRACING"`
	// EnableMetrics bool   `env:"TELEMETRY_ENABLE_METRICS"`
	ServiceName  string `env:"TELEMETRY_SERVICE_NAME"`
	OtlpEndpoint string `env:"TELEMETRY_OTLP_ENDPOINT"`
	// TraceSampleRatio    float64 `env:"TELEMETRY_TRACE_SAMPLE_RATIO"`
	// MetricsPushInterval int     `env:"TELEMETRY_METRICS_PUSH_INTERVAL"`
	// BetterStackEndpoint string  `env:"BETTERSTACK_ENDPOINT"`
	// BetterStackToken    string  `env:"BETTERSTACK_TOKEN"`
}

func NewConfig() Config {
	var cfg Config

	switch os.Getenv("ENVIRONMENT") {
	case DEV_ENVIRONMENT, PROD_ENVIRONMENT:
		awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		if awsAccessKeyID == "" {
			panic("missing 'AWS_ACCESS_KEY_ID'")
		}
		awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if awsSecretAccessKey == "" {
			panic("missing 'AWS_SECRET_ACCESS_KEY'")
		}

		cfg = Config{
			newDatabase(),
			newAuthentication(),
			newApp(),
			newTelemetry(),
			awsAccessKeyID,
			awsSecretAccessKey,
		}
	default:
		cfg = newTestConfig()
	}

	return cfg
}

func newTestConfig() Config {
	return Config{
		Authentication: Authentication{
			PasswordSalt:         "salty",
			SessionKey:           "session",
			SessionEncryptionKey: "session_enc_key",
			TokenSigningKey:      "token_signing_key",
			CsrfToken:            "csrf_token",
		},
		App: App{
			ServerHost:             "0.0.0.0",
			ServerPort:             "8080",
			AppDomain:              "testing",
			AppProtocol:            "http",
			ProjectName:            "test",
			Environment:            TEST_ENVIRONMENT,
			DefaultSenderSignature: "test@testing.com",
		},
		Telemetry: Telemetry{
			// EnableTracing:       true,
			// EnableMetrics:       true,
			ServiceName:  "grafto-test",
			OtlpEndpoint: "",
			// OtlpInsecure:        true,
			// TraceSampleRatio:    1.0,
			// MetricsPushInterval: 30,
		},
		AwsAccessKeyID:     "",
		AwsSecretAccessKey: "",
	}
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
