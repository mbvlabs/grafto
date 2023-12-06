package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/pages"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

type SignupPageData struct {
	PreviousNameInput  string
	PreviousEmailInput string
	Errors             validator.ValidationErrors
	WasSuccessful      bool
}

func SignupPage(ctx echo.Context, data SignupPageData) error {
	signup := pages.Signup{
		CsrfToken: csrf.Token(ctx.Request()),
	}

	containsErrors := len(data.Errors) > 0

	if !data.WasSuccessful && !containsErrors {
		return layouts.Base(signup.Page()).Render(extractRenderDeps(ctx))
	}

	if containsErrors {
		signup.NameInput.Value = data.PreviousNameInput
		signup.EmailInput.Value = data.PreviousEmailInput

		for _, validationError := range data.Errors {
			switch validationError.StructField() {
			case "Name":
				signup.NameInput = pages.TextInputData{ //TODO: not nice
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Mail":
				signup.EmailInput = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Password":
				signup.PasswordInput = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				signup.ConfirmPassword = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "MailRegistered":
				signup.EmailInput = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: "Email already registered",
				}
			}
		}

		return signup.Form().Render(extractRenderDeps(ctx))
	}

	return signup.Successful().Render(extractRenderDeps(ctx))
}

func VerifyEmail(ctx echo.Context, tokenInvalid bool) error {
	view := pages.VerifyEmail{
		TokenInvalid: tokenInvalid,
	}

	return layouts.Base(view.Page()).Render(extractRenderDeps(ctx))
}
