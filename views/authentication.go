package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/pages"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

type LoginPageData struct {
	CouldNotAuthenticate bool
	EmailNotVerified     bool
	WasSuccess           bool //TODO: not really too keen on this naming, revisit
}

func LoginPage(ctx echo.Context, data LoginPageData) error {
	login := pages.Login{
		EmailNotVerified:     data.EmailNotVerified,
		CouldNotAuthenticate: data.CouldNotAuthenticate,
		CsrfToken:            csrf.Token(ctx.Request()),
	}

	if !data.WasSuccess {
		return layouts.Base(login.Page()).Render(extractRenderDeps(ctx))
	}

	if data.CouldNotAuthenticate || data.EmailNotVerified {
		return layouts.Base(login.Form()).Render(extractRenderDeps(ctx))
	}

	return layouts.Base(login.Response()).Render(extractRenderDeps(ctx))
}

func ForgottenPasswordPage(ctx echo.Context, wasSuccess bool) error {
	forgottenPW := pages.ForgottenPassword{
		CsrfToken: csrf.Token(ctx.Request()),
	}

	if wasSuccess {
		layouts.Base(forgottenPW.Response()).Render(extractRenderDeps(ctx))
	}

	return layouts.Base(forgottenPW.Page()).Render(extractRenderDeps(ctx))
}

type ResetPasswordData struct {
	Token        string
	TokenInvalid bool
	Errors       validator.ValidationErrors
	WasSuccess   bool
}

func ResetPasswordPage(ctx echo.Context, data ResetPasswordData) error {
	resetPassword := pages.ResetPassword{
		CsrfToken:    csrf.Token(ctx.Request()),
		TokenInvalid: data.TokenInvalid,
		ResetToken:   data.Token,
	}

	if !data.WasSuccess {
		return layouts.Base(resetPassword.Page()).Render(extractRenderDeps(ctx))
	}

	if len(data.Errors) > 0 {
		for _, validationError := range data.Errors {
			switch validationError.StructField() {
			case "Password":
				resetPassword.Password = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				resetPassword.ConfirmPassword = pages.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			}
		}
		return layouts.Base(resetPassword.Form()).Render(extractRenderDeps(ctx))
	}

	return layouts.Base(resetPassword.Response()).Render(extractRenderDeps(ctx))
}
