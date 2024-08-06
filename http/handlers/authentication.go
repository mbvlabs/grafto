package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/validation"
	"github.com/mbv-labs/grafto/services"
	"github.com/mbv-labs/grafto/views"
	"github.com/mbv-labs/grafto/views/authentication"
)

type Authentication struct {
	Base
	authService services.Auth
	userModel   models.UserService
	tknService  services.Token
}

func NewAuthentication(
	authSvc services.Auth,
	base Base,
	userSvc models.UserService,
	tknManager services.Token,
) Authentication {
	return Authentication{base, authSvc, userSvc, tknManager}
}

func (a *Authentication) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(views.ExtractRenderDeps(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (a *Authentication) StoreAuthenticatedSession(ctx echo.Context) error {
	var payload StoreAuthenticatedSessionPayload
	if err := ctx.Bind(&payload); err != nil {
		telemetry.Logger.ErrorContext(
			ctx.Request().Context(),
			"could not parse UserLoginPayload",
			"error",
			err,
		)

		return authentication.LoginForm(csrf.Token(ctx.Request()), true, nil).
			Render(views.ExtractRenderDeps(ctx))
	}

	if err := a.authService.AuthenticateUser(
		ctx.Request().Context(),
		payload.Mail,
		payload.Password,
	); err != nil {
		telemetry.Logger.ErrorContext(
			ctx.Request().Context(),
			"could not authenticate user",
			"error",
			err,
		)

		errors := make(views.Errors)

		switch err {
		case services.ErrPasswordNotMatch, services.ErrUserNotExist:
			errors[authentication.ErrAuthDetailsWrong] = "The email or password you entered is incorrect."
		case services.ErrEmailNotValidated:
			errors[authentication.ErrEmailNotValidated] = "Your email has not yet been verified."
		}

		return authentication.LoginForm(csrf.Token(ctx.Request()), false, errors).
			Render(views.ExtractRenderDeps(ctx))
	}

	user, err := a.userModel.ByEmail(ctx.Request().Context(), payload.Mail)
	if err != nil {
		return a.InternalError(ctx)
	}

	_, err = a.authService.NewUserSession(ctx.Request(), ctx.Response(), user.ID)
	if err != nil {
		return err
	}

	return authentication.LoginForm(csrf.Token(ctx.Request()), true, nil).
		Render(views.ExtractRenderDeps(ctx))
}

func (a *Authentication) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(csrf.Token(ctx.Request())).
		Render(views.ExtractRenderDeps(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a *Authentication) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).Render(views.ExtractRenderDeps(ctx))
	}

	user, err := a.db.QueryUserByEmail(ctx.Request().Context(), payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
				CsrfToken:        csrf.Token(ctx.Request()),
				NoAssociatedUser: true,
			}).Render(views.ExtractRenderDeps(ctx))
		}

		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).Render(views.ExtractRenderDeps(ctx))
	}
	resetToken, err := a.tknService.CreateResetPasswordToken(ctx.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	// TODO fix this error flow
	pwResetMail := &templates.PasswordResetMail{
		ResetPasswordLink: fmt.Sprintf(
			"%s://%s/reset-password?token=%s",
			a.cfg.App.AppScheme,
			a.cfg.App.AppHost,
			resetToken,
		),
	}

	textVersion, err := pwResetMail.GenerateTextVersion()
	if err != nil {
		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).Render(views.ExtractRenderDeps(ctx))
	}
	htmlVersion, err := pwResetMail.GenerateHtmlVersion()
	if err != nil {
		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).Render(views.ExtractRenderDeps(ctx))
	}

	_, err = a.queueClient.Insert(ctx.Request().Context(), queue.EmailJobArgs{
		To:          user.Mail,
		From:        a.cfg.App.DefaultSenderSignature,
		Subject:     "Password Reset Request",
		TextVersion: textVersion,
		HtmlVersion: htmlVersion,
	}, nil)
	if err != nil {
		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
		Success:   true,
	}).Render(views.ExtractRenderDeps(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

func (a *Authentication) CreateResetPassword(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return authentication.ResetPasswordPage(false, true, csrf.Token(ctx.Request()), "").
			Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.ResetPasswordPage(false, false, csrf.Token(ctx.Request()), passwordResetToken.Token).
		Render(views.ExtractRenderDeps(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

func (a *Authentication) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ResetPasswordPage(false, true, "", "").
			Render(views.ExtractRenderDeps(ctx))
	}

	if err := a.tknService.Validate(ctx.Request().Context(), payload.Token, services.ScopeResetPassword); err != nil {
		return err
	}

	userID, err := a.tknService.GetAssociatedUserID(ctx.Request().Context(), payload.Token)
	if err != nil {
		return authentication.ResetPasswordPage(false, true, "", "").
			Render(views.ExtractRenderDeps(ctx))
	}

	err = a.userModel.ChangePassword(ctx.Request().Context(),
		models.ChangeUserPasswordData{
			ID:              userID,
			UpdatedAt:       time.Now(),
			Password:        payload.Password,
			ConfirmPassword: payload.ConfirmPassword,
		},
	)
	if err != nil && errors.Is(err, models.ErrFailValidation) {
		var valiErrs validation.ValidationErrors
		if ok := errors.As(err, &valiErrs); !ok {
			return a.InternalError(ctx)
		}

		props := authentication.ResetPasswordFormProps{
			CsrfToken:  csrf.Token(ctx.Request()),
			ResetToken: payload.Token,
		}

		for _, validationError := range valiErrs {
			switch validationError.GetFieldName() {
			case "Password":
				props.Errors[authentication.PasswordField] = validationError.GetHumanExplanations()[0]
			case "ConfirmPassword":
				props.Errors[authentication.PasswordField] = validationError.GetHumanExplanations()[0]
			}
		}

		return authentication.ResetPasswordForm(props).Render(views.ExtractRenderDeps(ctx))
	}
	if err != nil {
		return a.InternalError(ctx)
	}

	if err := a.tknService.Delete(ctx.Request().Context(), payload.Token); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return a.InternalError(ctx)
	}

	return authentication.ResetPasswordForm(authentication.ResetPasswordFormProps{}).
		Render(views.ExtractRenderDeps(ctx))
}
