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

func NewTBD() Config {
	db := newDatabase()
	authentication := newAuthentication()
	app := newApp()
	telemetry := newTelemetry()

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		panic("missing 'AWS_ACCESS_KEY_ID'")
	}
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		panic("missing 'AWS_SECRET_ACCESS_KEY'")
	}

	return Config{
		db,
		authentication,
		app,
		telemetry,
		awsAccessKeyID,
		awsSecretAccessKey,
	}
}
