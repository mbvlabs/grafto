package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type config struct {
	database       configDatabase
	authentication configAuthentication
}

var Cfg config = setupConfiguration()

func setupConfiguration() config {
	databaseCfg := configDatabase{}
	if err := env.Parse(&databaseCfg); err != nil {
		panic(err)
	}

	authCfg := configAuthentication{}
	if err := env.Parse(&authCfg); err != nil {
		panic(err)
	}
	return config{
		databaseCfg,
		authCfg,
	}
}

func (c config) GetDatabaseURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		c.database.DatabaseKind, c.database.User, c.database.Password, c.database.Host, c.database.Port,
		c.database.Name, c.database.SSL_MODE,
	)
}

func (c config) GetPwdPepper() string {
	return c.authentication.pwdPepper
}
