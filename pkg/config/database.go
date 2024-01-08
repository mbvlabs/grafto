package config

type configDatabase struct {
	Port         string `env:"DB_PORT"`
	Host         string `env:"DB_HOST"`
	Name         string `env:"DB_NAME"`
	User         string `env:"DB_USER"`
	Password     string `env:"DB_PASSWORD"`
	DatabaseKind string `env:"DB_KIND"`
	SSL_MODE     string `env:"DB_SSL_MODE"`
}

// func GetDatabaseURL() string {
// 	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
// 		config.database.databaseKind, config.database.user, config.database.password, config.database.host, config.database.port,
// 		config.database.name,
// 	)
// }
