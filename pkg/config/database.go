package config

import "fmt"

type configDatabase struct {
	port         string `env:"PORT"`
	host         string `env:"HOST"`
	name         string `env:"NAME"`
	user         string `env:"USER"`
	password     string `env:"PASSWORD"`
	databaseKind string `env:"DATABASE_KIND"`
}

func GetDatabaseURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		config.database.databaseKind, config.database.user, config.database.password, config.database.host, config.database.port,
		config.database.name,
	)
}
