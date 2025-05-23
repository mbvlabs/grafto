package handlers

import (
	"errors"
	"log/slog"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/routes"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
)

type Authentication struct {
	db          psql.Postgres
	emailClient services.EmailSender
}

func newAuthentication(
	db psql.Postgres,
	emailClient services.EmailSender,
) Authentication {
	return Authentication{db, emailClient}
}

func (a Authentication) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (a Authentication) StoreAuthenticatedSession(ctx echo.Context) error {
	var payload StoreAuthenticatedSessionPayload
	if err := ctx.Bind(&payload); err != nil {
		slog.ErrorContext(
			ctx.Request().Context(),
			"could not parse UserLoginPayload",
			"error",
			err,
		)

		return views.ErrorPage().Render(renderArgs(ctx))
	}

	authenticatedUser, err := services.AuthenticateUser(
		ctx.Request().Context(),
		a.db,
		payload.Mail,
		payload.Password,
	)
	if err != nil {
		var userErr views.Errors
		if errors.Is(err, services.ErrUserEmailNotVerified) {
			userErr = views.Errors{
				authentication.ErrEmailNotValidated: "Your email has not yet been verified.",
			}
		}
		if errors.Is(err, services.ErrInvalidAuthDetail) {
			userErr = views.Errors{
				authentication.ErrEmailNotValidated: "The email or password you entered is incorrect.",
			}
		}

		return authentication.LoginForm(
			csrf.Token(
				ctx.Request(),
			),
			false,
			userErr,
		).
			Render(renderArgs(ctx))
	}

	if err := createAuthSession(
		ctx, payload.RememberMe == "on", authenticatedUser); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.LoginForm(
		csrf.Token(ctx.Request()), true, nil).
		Render(renderArgs(ctx))
}

func (a Authentication) DestroyAuthenticatedSession(ctx echo.Context) error {
	if err := destroyAuthSession(ctx); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return redirect(
		ctx.Response(),
		ctx.Request(),
		routes.LoginPage.Path,
	)
}

func (a Authentication) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(csrf.Token(ctx.Request())).
		Render(renderArgs(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a Authentication) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.SendResetPasswordEmail(ctx.Request().Context(), a.db, a.emailClient, payload.Email); err != nil {
		// TODO: show proper error page with info
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
		Success:   true,
	}).
		Render(renderArgs(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

func (a Authentication) CreateResetPassword(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.ResetPasswordPage(
		false, false, csrf.Token(ctx.Request()), passwordResetToken.Token).
		Render(renderArgs(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

func (a Authentication) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.ChangeUserPassword(ctx.Request().Context(), a.db, payload.Token, payload.Password, payload.ConfirmPassword); err != nil {
		// TODO: show proper error page with info
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.ResetPasswordForm(authentication.ResetPasswordFormProps{}).
		Render(renderArgs(ctx))
}
