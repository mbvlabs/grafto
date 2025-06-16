package handlers

import (
	"log/slog"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/fragments"
	"github.com/mbvlabs/grafto/views/sessions"
)

type Registrations struct {
	db          psql.Postgres
	emailClient services.EmailSender
}

func newRegistrations(
	db psql.Postgres,
	emailClient services.EmailSender,
) Registrations {
	return Registrations{db, emailClient}
}

// New renders the registration page (GET /registrations/new)
func (r Registrations) New(ctx echo.Context) error {
	return sessions.RegisterPage(sessions.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreUserPayload struct {
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

// Create handles user registration (POST /registrations)
func (r Registrations) Create(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.RegisterUser(
		ctx.Request().Context(), r.db, r.emailClient, payload.Email, payload.Password, payload.ConfirmPassword); err != nil {
		slog.InfoContext(
			ctx.Request().Context(),
			"could not register user",
			"err",
			err,
		)

		// TODO handle err
		return err
	}

	return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
		CsrfToken:   csrf.Token(ctx.Request()),
		CodeInvalid: false,
		Success:     false,
	}).Render(renderArgs(ctx))
}

type verificationCodePayload struct {
	Code string `form:"code"`
}

// Update handles email verification (POST /verify-email)
func (r Registrations) Update(ctx echo.Context) error {
	var payload verificationCodePayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.ValidateUserEmail(
		ctx.Request().Context(),
		r.db,
		payload.Code,
	); err != nil {
		return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
			CsrfToken:   csrf.Token(ctx.Request()),
			CodeInvalid: true,
			Success:     false,
		}).Render(renderArgs(ctx))
	}

	return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
		CsrfToken:   csrf.Token(ctx.Request()),
		CodeInvalid: false,
		Success:     true,
	}).Render(renderArgs(ctx))
}
