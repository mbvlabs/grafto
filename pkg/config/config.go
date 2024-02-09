package config

import (
	"github.com/caarlos0/env/v10"
)

type Cfg struct {
	Db                Database
	Auth              Authentication
	App               App
	ExternalProviders ExternalProviders
}

func New() Cfg {
	databaseCfg := Database{}
	if err := env.Parse(&databaseCfg); err != nil {
		panic(err)
	}

	authCfg := Authentication{}
	if err := env.Parse(&authCfg); err != nil {
		panic(err)
	}

	appCfg := App{}
	if err := env.Parse(&appCfg); err != nil {
		panic(err)
	}

	externalProviders := ExternalProviders{}
	if err := env.Parse(&externalProviders); err != nil {
		panic(err)
	}

	return Cfg{
		databaseCfg,
		authCfg,
		appCfg,
		externalProviders,
	}
}
