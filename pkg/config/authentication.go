package config

type Authentication struct {
	PasswordPepper       string `env:"PASSWORD_PEPPER"`
	SessionKey           string `env:"SESSION_KEY"`
	SessionEncryptionKey string `env:"SESSION_ENCRYPTION_KEY"`
	TokenSigningKey      string `env:"TOKEN_SIGNING_KEY"`
	CsrfToken            string `env:"CSRF_TOKEN"`
}
