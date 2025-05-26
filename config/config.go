package config

import (
	"os"
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
			ServiceName:  "grafto-test",
			OtlpEndpoint: "",
		},
		AwsAccessKeyID:     "",
		AwsSecretAccessKey: "",
	}
}
