package views

import (
	"github.com/MBvisti/grafto/pkg/telemetry"
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

type ForgottenPasswordPageData struct {
	RenderSuccessResponse bool
}

func ForgottenPasswordPage(ctx echo.Context, data ForgottenPasswordPageData) error {
	forgottenPW := pages.ForgottenPassword{
		CsrfToken: csrf.Token(ctx.Request()),
	}

	if data.RenderSuccessResponse {
		return forgottenPW.Response().Render(extractRenderDeps(ctx))
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

	hasErrors := len(data.Errors) > 0

	if !data.WasSuccess && !hasErrors && !data.TokenInvalid {
		return layouts.Base(resetPassword.Page()).Render(extractRenderDeps(ctx))
	}

	telemetry.Logger.Info("status", "success", data.WasSuccess, "errors", hasErrors, "token", data.TokenInvalid)
	if !data.WasSuccess && !hasErrors && data.TokenInvalid {
		return resetPassword.Page().Render(extractRenderDeps(ctx))
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
		return resetPassword.Page().Render(extractRenderDeps(ctx))
	}

	return resetPassword.Response().Render(extractRenderDeps(ctx))
}
