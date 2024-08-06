package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

const (
	DEV_ENVIRONMENT  = "development"
	PROD_ENVIRONMENT = "production"
)

type App struct {
	ServerHost             string `env:"SERVER_HOST"`
	ServerPort             string `env:"SERVER_PORT"`
	AppDomain              string `env:"APP_DOMAIN"`
	AppProtocol            string `env:"APP_PROTOCOL"`
	ProjectName            string `env:"PROJECT_NAME"`
	Environment            string `env:"ENVIRONMENT"`
	DefaultSenderSignature string `env:"DEFAULT_SENDER_SIGNATURE"`
}

func (a App) GetFullDomain() string {
	return fmt.Sprintf("%v://%v", a.AppProtocol, a.AppDomain)
}

func newApp() App {
	appCfg := App{}

	if err := env.Parse(&appCfg); err != nil {
		panic(err)
	}

	return appCfg
}
