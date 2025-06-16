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
	sessions "github.com/mbvlabs/grafto/views/sessions"
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

// New renders the login page (GET /sessions/new)
func (a Sessions) New(ctx echo.Context) error {
	return sessions.LoginPage(sessions.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

// Create handles session creation (POST /sessions)
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
		payload.Mail,
		payload.Password,
	)
	if err != nil {
		var userErr views.Errors
		if errors.Is(err, services.ErrUserEmailNotVerified) {
			userErr = views.Errors{
				sessions.ErrEmailNotValidated: "Your email has not yet been verified.",
			}
		}
		if errors.Is(err, services.ErrInvalidAuthDetail) {
			userErr = views.Errors{
				sessions.ErrEmailNotValidated: "The email or password you entered is incorrect.",
			}
		}

		return sessions.LoginForm(
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

	return sessions.LoginForm(
		csrf.Token(ctx.Request()), true, nil).
		Render(renderArgs(ctx))
}

// Destroy handles session destruction (DELETE /sessions)
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

// NewPasswordReset renders the password reset request page (GET /password-resets/new)
func (a Sessions) NewPasswordReset(ctx echo.Context) error {
	return sessions.ForgottenPasswordPage(csrf.Token(ctx.Request())).
		Render(renderArgs(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

// CreatePasswordReset handles password reset request creation (POST /password-resets)
func (a Sessions) CreatePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.SendResetPasswordEmail(ctx.Request().Context(), a.db, a.emailClient, payload.Email); err != nil {
		// TODO: show proper error page with info
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return sessions.ForgottenPasswordForm(sessions.ForgottenPasswordFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
		Success:   true,
	}).
		Render(renderArgs(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

// EditPasswordReset renders the password reset form (GET /password-resets/edit)
func (a Sessions) EditPasswordReset(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return sessions.ResetPasswordPage(
		false, false, csrf.Token(ctx.Request()), passwordResetToken.Token).
		Render(renderArgs(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

// UpdatePasswordReset handles password reset form submission (PUT /password-resets)
func (a Sessions) UpdatePasswordReset(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.ChangeUserPassword(ctx.Request().Context(), a.db, payload.Token, payload.Password, payload.ConfirmPassword); err != nil {
		// TODO: show proper error page with info
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return sessions.ResetPasswordForm(sessions.ResetPasswordFormProps{}).
		Render(renderArgs(ctx))
}
