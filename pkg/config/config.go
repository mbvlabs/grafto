package config

type Config struct {
	database       configDatabase
	authentication configAuthentication
}

var config Config = setupConfiguration()

func setupConfiguration() Config {
	return Config{}
}
