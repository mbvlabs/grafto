package config

type App struct {
	ServerHost             string `env:"SERVER_HOST"`
	ServerPort             string `env:"SERVER_PORT"`
	AppHost                string `env:"APP_HOST"`
	AppScheme              string `env:"APP_SCHEME"`
	ProjectName            string `env:"PROJECT_NAME"`
	Environment            string `env:"ENVIRONMENT"`
	DefaultSenderSignature string `env:"DEFAULT_SENDER_SIGNATURE"`
}
