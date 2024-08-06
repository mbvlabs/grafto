package config

import "os"

type TBD struct {
	Database
	Authentication
	App
	AwsAccessKeyID     string
	AwsSecretAccessKey string
}

func NewTBD() TBD {
	db := newDatabase()
	authentication := newAuthentication()
	app := newApp()

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		panic("missing 'AWS_ACCESS_KEY_ID'")
	}
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		panic("missing 'AWS_SECRET_ACCESS_KEY'")
	}

	return TBD{
		db,
		authentication,
		app,
		awsAccessKeyID,
		awsSecretAccessKey,
	}
}
