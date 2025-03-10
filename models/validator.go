package models

import (
	"github.com/go-playground/validator/v10"
)

var validate = setupValidator()

func setupValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterStructValidation(validatePasswordsMatch, PasswordPair{})

	return v
}

func validatePasswordsMatch(sl validator.StructLevel) {
	pwPair := sl.Current().Interface().(PasswordPair)

	if pwPair.Password != pwPair.ConfirmPassword {
		sl.ReportError(
			pwPair.Password,
			"Password",
			"Password",
			"must match confirm password",
			"",
		)
		sl.ReportError(
			pwPair.Password,
			"ConfirmPassword",
			"ConfirmPassword",
			"must match password",
			"",
		)
	}
}
