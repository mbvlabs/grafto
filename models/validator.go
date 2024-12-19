package models

import (
	"github.com/go-playground/validator/v10"
)

var validate = setupValidator()

func setupValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterStructValidation(validateConfirmPWMatch, NewUserPayload{})

	return v
}

func validateConfirmPWMatch(sl validator.StructLevel) {
	user := sl.Current().Interface().(NewUserPayload)

	if user.Password != user.ConfirmPassword {
		sl.ReportError(user.Password, "Password", "Password", "must match confirm password", "")
		sl.ReportError(user.Password, "ConfirmPassword", "ConfirmPassword", "must match password", "")
	}
}
