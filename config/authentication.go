package config

import "github.com/caarlos0/env/v10"

type Authentication struct {
	PasswordPepper       string `env:"PASSWORD_PEPPER"`
	SessionKey           string `env:"SESSION_KEY"`
	SessionEncryptionKey string `env:"SESSION_ENCRYPTION_KEY"`
	TokenSigningKey      string `env:"TOKEN_SIGNING_KEY"`
	CsrfToken            string `env:"CSRF_TOKEN"`
}

func newAuthentication() Authentication {
	authenticationCfg := Authentication{}

	if err := env.Parse(&authenticationCfg); err != nil {
		panic(err)
	}

	return authenticationCfg
}
