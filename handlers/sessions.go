package handlers

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/routes"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	sessionViews "github.com/mbvlabs/grafto/views/sessions"
)

type Sessions struct {
	db          psql.Postgres
	emailClient services.EmailSender
}

func newSessions(
	db psql.Postgres,
	emailClient services.EmailSender,
) Sessions {
	return Sessions{db, emailClient}
}

func (a Sessions) New(ctx echo.Context) error {
	return sessionViews.LoginPage(sessionViews.LoginPageProps{}).
		Render(renderArgs(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Email      string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (a Sessions) Create(ctx echo.Context) error {
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
		payload.Email,
		payload.Password,
	)
	if err != nil {
		var userErr views.Errors
		if errors.Is(err, services.ErrUserEmailNotVerified) {
			userErr = views.Errors{
				sessionViews.ErrEmailNotValidated: "Your email has not yet been verified.",
			}
		}
		if errors.Is(err, services.ErrInvalidAuthDetail) {
			userErr = views.Errors{
				sessionViews.ErrEmailNotValidated: "The email or password you entered is incorrect.",
			}
		}

		return sessionViews.LoginForm(false, userErr).Render(renderArgs(ctx))
	}

	if err := createAuthSession(
		ctx, payload.RememberMe == "on", authenticatedUser); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return sessionViews.LoginForm(true, views.Errors{}).Render(renderArgs(ctx))
}

func (a Sessions) Destroy(ctx echo.Context) error {
	if err := destroyAuthSession(ctx); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return redirect(
		ctx.Response(),
		ctx.Request(),
		routes.LoginPage.Path,
	)
}

func (a Sessions) NewPasswordReset(ctx echo.Context) error {
	return sessionViews.ForgottenPasswordPage().
		Render(renderArgs(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a Sessions) CreatePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return sessionViews.ForgottenPasswordForm(
			sessionViews.ForgottenPasswordFormProps{HasError: true},
		).Render(renderArgs(ctx))
	}

	if err := services.SendResetPasswordEmail(ctx.Request().Context(), a.db, a.emailClient, payload.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sessionViews.ForgottenPasswordForm(
				sessionViews.ForgottenPasswordFormProps{Success: true},
			).Render(renderArgs(ctx))
		}

		return sessionViews.ForgottenPasswordForm(
			sessionViews.ForgottenPasswordFormProps{HasError: true},
		).Render(renderArgs(ctx))
	}

	return sessionViews.ForgottenPasswordForm(
		sessionViews.ForgottenPasswordFormProps{Success: true},
	).Render(renderArgs(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

func (a Sessions) EditPasswordReset(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return sessionViews.ResetPasswordPage(
		false, false, csrf.Token(ctx.Request()), passwordResetToken.Token).
		Render(renderArgs(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

func (a Sessions) UpdatePasswordReset(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return sessionViews.ResetPasswordForm(sessionViews.ResetPasswordFormProps{Errors: views.Errors{"request": "could not bind"}}).
			Render(renderArgs(ctx))
	}

	if err := services.ChangeUserPassword(ctx.Request().Context(), a.db, payload.Token, payload.Password, payload.ConfirmPassword); err != nil {
		if errors.Is(err, models.ErrDomainValidation) {
			return sessionViews.ResetPasswordForm(sessionViews.ResetPasswordFormProps{Errors: views.Errors{"password": err.Error(), "confirm_password": err.Error()}}).
				Render(renderArgs(ctx))
		}

		return sessionViews.ResetPasswordForm(sessionViews.ResetPasswordFormProps{Errors: views.Errors{"request": "internal error"}}).
			Render(renderArgs(ctx))
	}

	return sessionViews.ResetPasswordForm(sessionViews.ResetPasswordFormProps{Success: true}).
		Render(renderArgs(ctx))
}
