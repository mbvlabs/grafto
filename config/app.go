package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

const (
	DEV_ENVIRONMENT  = "development"
	TEST_ENVIRONMENT = "testing"
	PROD_ENVIRONMENT = "production"
)

type Application struct {
	ServerHost             string `env:"SERVER_HOST"`
	ServerPort             string `env:"SERVER_PORT"`
	Domain                 string `env:"APP_DOMAIN"`
	Protocol               string `env:"APP_PROTOCOL"`
	ProjectName            string `env:"PROJECT_NAME"`
	Env                    string `env:"ENVIRONMENT"`
	DefaultSenderSignature string `env:"DEFAULT_SENDER_SIGNATURE"`
}

func (a Application) GetFullDomain() string {
	if a.Env == DEV_ENVIRONMENT {
		return fmt.Sprintf(
			"%v://%v:%v",
			a.Protocol,
			a.Domain,
			a.ServerPort,
		)
	}
	return fmt.Sprintf("%v://%v", a.Protocol, a.Domain)
}

func newApp() Application {
	appCfg := Application{}

	if err := env.ParseWithOptions(&appCfg, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}

	return appCfg
}
