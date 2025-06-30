package config

import (
	"os"
)

// Cfg instantiate a new cfg but can panic
var Cfg Config = NewConfig()

type Config struct {
	DB                 Database
	Auth               Authentication
	App                Application
	Telemetry          Telemetry
	AwsAccessKeyID     string
	AwsSecretAccessKey string
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
		Auth: Authentication{
			PasswordSalt:         "salty",
			SessionKey:           "44e72c7a3745db47d88c07b82e16917c6ca6ce30f2930fa2f67fec0fe001bb06469d61f5b3de80828d647f29e6b2ac7f40ff9f5b792dbb4a18c3b0420ed7343c",
			SessionEncryptionKey: "5d32a679df4834fac7c919c06cf37162e73287e38ae6151fc054abac9ab40dfc",
			TokenSigningKey:      "573ec461fea4cad049ffdcf0d08c353472f24adfec64529dfcafd044bee1f442",
		},
		App: Application{
			ServerHost:             "0.0.0.0",
			ServerPort:             "8080",
			Domain:                 "testing",
			Protocol:               "http",
			ProjectName:            "test",
			Env:                    TEST_ENVIRONMENT,
			DefaultSenderSignature: "test@testing.com",
		},
		Telemetry: Telemetry{
			ServiceName:  "grafto-test",
			OtlpEndpoint: "",
		},
		AwsAccessKeyID:     "",
		AwsSecretAccessKey: "",
	}
}
