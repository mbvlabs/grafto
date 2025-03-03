package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/emails"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
	"github.com/mbvlabs/grafto/views/paths"
)

type Authentication struct {
	db       psql.Postgres
	emailSvc EmailService
}

func newAuthentication(
	db psql.Postgres,
	emailSvc EmailService,
) Authentication {
	return Authentication{db, emailSvc}
}

func (a *Authentication) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (a *Authentication) StoreAuthenticatedSession(ctx echo.Context) error {
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

	user, err := models.GetUserByEmail(
		ctx.Request().Context(),
		payload.Mail,
		a.db.Pool,
	)
	if err != nil {
		return authentication.LoginForm(
			csrf.Token(
				ctx.Request(),
			),
			false,
			views.Errors{
				authentication.ErrEmailNotValidated: "The email or password you entered is incorrect.",
			},
		).
			Render(renderArgs(ctx))
	}

	if !user.IsVerified() {
		return authentication.LoginForm(
			csrf.Token(
				ctx.Request(),
			),
			false,
			views.Errors{
				authentication.ErrEmailNotValidated: "Your email has not yet been verified.",
			},
		).
			Render(renderArgs(ctx))
	}

	if err := user.ValidatePassword(
		payload.Password); err != nil {
		return authentication.LoginForm(
			csrf.Token(
				ctx.Request(),
			),
			false,
			views.Errors{
				authentication.ErrAuthDetailsWrong: "The email or password you entered is incorrect.",
			},
		).
			Render(renderArgs(ctx))
	}

	if err := createAuthSession(
		ctx, payload.RememberMe == "on", user); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.LoginForm(
		csrf.Token(ctx.Request()), true, nil).
		Render(renderArgs(ctx))
}

func (a *Authentication) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(csrf.Token(ctx.Request())).
		Render(renderArgs(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a *Authentication) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	user, err := models.GetUserByEmail(
		ctx.Request().Context(),
		payload.Email,
		a.db.Pool,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	tkn, err := models.NewToken(
		ctx.Request().Context(),
		a.db.Pool,
		models.NewTokenPayload{
			Expiration: models.ResetPasswordExpirary,
			Meta: models.MetaInformation{
				Resource:   models.ResourceUser,
				ResourceID: user.ID,
				Scope:      models.ScopeResetPassword,
			},
		},
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	html, txt, err := emails.PasswordReset{
		ResetLink: fmt.Sprintf(
			"%s/%s?token=%s",
			config.Cfg.GetFullDomain(),
			paths.Get(ctx.Request().Context(), paths.ResetPasswordPage),
			tkn.Hash,
		),
	}.Generate(ctx.Request().Context())
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := a.emailSvc.Send(ctx.Request().Context(), services.EmailPayload{
		To:       user.Email,
		Subject:  "Action Required | Password reset requested",
		HtmlBody: html.String(),
		TextBody: txt.String(),
	}); err != nil {
		return err
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

func (a *Authentication) CreateResetPassword(ctx echo.Context) error {
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

func (a *Authentication) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	token, err := models.GetToken(
		ctx.Request().Context(),
		a.db.Pool,
		payload.Token,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if !token.IsValid() || token.Meta.Scope != models.ScopeResetPassword {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := models.UpdateUserPassword(
		ctx.Request().Context(),
		a.db.Pool,
		models.UpdateUserPasswordPayload{
			ID:        token.Meta.ResourceID,
			UpdatedAt: time.Now(),
			Password: models.PasswordPair{
				Password:        payload.Password,
				ConfirmPassword: payload.ConfirmPassword,
			},
		},
	); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := models.DeleteToken(
		ctx.Request().Context(), a.db.Pool, token.ID); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return authentication.ResetPasswordForm(authentication.ResetPasswordFormProps{}).
		Render(renderArgs(ctx))
}
