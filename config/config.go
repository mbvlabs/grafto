package config

import "os"

type Config struct {
	Database
	Authentication
	App
	Telemetry
	AwsAccessKeyID     string
	AwsSecretAccessKey string
}

func NewConfig() Config {
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		panic("missing 'AWS_ACCESS_KEY_ID'")
	}
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		panic("missing 'AWS_SECRET_ACCESS_KEY'")
	}

	return Config{
		newDatabase(),
		newAuthentication(),
		newApp(),
		newTelemetry(),
		awsAccessKeyID,
		awsSecretAccessKey,
	}
}
