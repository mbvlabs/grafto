package handlers

import (
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
)

type Registration struct {
	db          psql.Postgres
	emailClient EmailClient
}

func newRegistration(
	db psql.Postgres,
	emailClient EmailClient,
) Registration {
	return Registration{db, emailClient}
}

func (r *Registration) CreateUser(ctx echo.Context) error {
	return authentication.RegisterPage(authentication.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreUserPayload struct {
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

// TODO: send email validation email
func (r *Registration) StoreUser(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if _, err := models.NewUser(ctx.Request().Context(), models.NewUserPayload{
		Email: payload.Email,
		Password: models.PasswordPair{
			Password:        payload.Password,
			ConfirmPassword: payload.ConfirmPassword,
		},
	}, r.db.Pool); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	props := authentication.RegisterFormProps{
		SuccessRegister: true,
		CsrfToken:       csrf.Token(ctx.Request()),
	}
	return authentication.RegisterForm(props).
		Render(renderArgs(ctx))
}

type verificationTokenPayload struct {
	Token string `query:"token"`
}

func (r *Registration) VerifyUserEmail(ctx echo.Context) error {
	var payload verificationTokenPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	token, err := models.GetToken(
		ctx.Request().Context(),
		r.db.Pool,
		payload.Token,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if !token.IsValid() || token.Meta.Scope != models.ScopeEmailVerification {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	user, err := models.GetUser(
		ctx.Request().Context(),
		token.Meta.ResourceID,
		r.db.Pool,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := models.UpdateUserEmailToVerified(
		ctx.Request().Context(),
		models.UpdateUserEmailToVerifiedPayload{
			ID:         user.ID,
			Email:      user.Email,
			VerifiedAt: time.Now(),
		},
		r.db.Pool,
	); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	// TODO: create auth session
	return authentication.VerifyEmailPage(false).
		Render(renderArgs(ctx))
}
