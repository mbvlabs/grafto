package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/templates"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

type SignupData struct {
	PreviousNameInput  string
	PreviousEmailInput string
	Errors             validator.ValidationErrors
	RenderPartial      bool
}

func Signup(ctx echo.Context, data SignupData) error {
	templateData := templates.RegisterUserData{CsrfToken: csrf.Token(ctx.Request())}

	if len(data.Errors) == 0 && data.RenderPartial {
		return templates.RegisterUserForm(templateData).Render(extractRenderDeps(ctx))
	}

	if len(data.Errors) > 0 {
		templateData.NameInput.Value = data.PreviousNameInput
		templateData.EmailInput.Value = data.PreviousEmailInput

		for _, validationError := range data.Errors {
			switch validationError.StructField() {
			case "Name":
				templateData.NameInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Mail":
				templateData.EmailInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Password":
				templateData.PasswordInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				templateData.ConfirmPassword = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "MailRegistered":
				templateData.EmailInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: "Email already registered",
				}
			}
		}

		return templates.RegisterUserForm(templateData).Render(extractRenderDeps(ctx))
	}

	return layouts.Base(templates.UserRegistrationWrapper(templateData)).Render(extractRenderDeps(ctx))
}

func SignupResponse(ctx echo.Context) error {
	return templates.UserRegisteredResponse().Render(extractRenderDeps(ctx))
}

func ForgottenPassword(ctx echo.Context) error {
	templateData := templates.ForgottenPasswordFormData{
		CsrfToken: csrf.Token(ctx.Request()),
	}

	return layouts.Base(templates.ForgottenPassword(templateData)).Render(extractRenderDeps(ctx))
}

func ForgottenPasswordResponse(ctx echo.Context) error {
	return templates.ForgottenPasswordResponse().Render(extractRenderDeps(ctx))
}

type ResetPasswordData struct {
	Token        string
	TokenInvalid bool
	Errors       validator.ValidationErrors
}

func ResetPassword(ctx echo.Context, data ResetPasswordData) error {
	templateData := templates.ResetPasswordFormData{
		CsrfToken:    csrf.Token(ctx.Request()),
		TokenInvalid: data.TokenInvalid,
	}

	if len(data.Errors) > 0 {
		for _, validationError := range data.Errors {
			switch validationError.StructField() {
			case "Password":
				templateData.Password = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				templateData.ConfirmPassword = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			}
		}
	}

	return layouts.Base(templates.ResetPassword(templateData)).Render(extractRenderDeps(ctx))
}

func ResetPasswordResponse(ctx echo.Context) error {
	return templates.ResetPasswordResponse().Render(extractRenderDeps(ctx))
}
