package services

import "github.com/go-playground/validator/v10"

func resetPasswordMatchValidation(sl validator.StructLevel) {
	data := sl.Current().Interface().(updateUserValidation)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(data.ConfirmPassword, "", "ConfirmPassword", "", "confirm password must match password")
	}
}

func passwordMatchValidation(sl validator.StructLevel) {
	data := sl.Current().Interface().(newUserValidation)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(data.ConfirmPassword, "", "ConfirmPassword", "", "confirm password must match password")
	}
}

func registerStructValidations(v *validator.Validate) *validator.Validate {
	v.RegisterStructValidation(resetPasswordMatchValidation, updateUserValidation{})
	v.RegisterStructValidation(passwordMatchValidation, newUserValidation{})

	return v
}
