package config

type configAuthentication struct {
	pwdPepper string `env:"PASSWORD_PEPPER"`
}

func GetPwdPepper() string {
	return config.authentication.pwdPepper
}
