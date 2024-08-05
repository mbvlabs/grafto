package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/pkg/validation"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/mbv-labs/grafto/services"
	"github.com/mbv-labs/grafto/views"
	"github.com/mbv-labs/grafto/views/authentication"
)

type Authentication struct {
	Base
	authService services.Auth
	userModel   models.UserService
	tknManager  tokens.Manager
}

func NewAuthentication(
	authSvc services.Auth,
	base Base,
	userSvc models.UserService,
	tknManager tokens.Manager,
) Authentication {
	return Authentication{base, authSvc, userSvc, tknManager}
}

func (a *Authentication) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
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

		return authentication.LoginResponse().Render(views.ExtractRenderDeps(ctx))
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

		var errors views.Errors

		switch err {
		case services.ErrPasswordNotMatch, services.ErrUserNotExist:
			errors[authentication.ErrAuthDetailsWrong] = "The email or password you entered is incorrect."
		case services.ErrEmailNotValidated:
			errors[authentication.ErrEmailNotValidated] = "Your email has not yet been verified."
		}

		return authentication.LoginForm(csrf.Token(ctx.Request()), errors).
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

	return authentication.LoginResponse().Render(views.ExtractRenderDeps(ctx))
}

func (a *Authentication) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(authentication.ForgottenPasswordPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a *Authentication) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	user, err := a.db.QueryUserByEmail(ctx.Request().Context(), payload.Email)
	if err != nil {
		failureOccurred := true
		if errors.Is(err, pgx.ErrNoRows) {
			failureOccurred = false
		}

		return authentication.ForgottenPasswordSuccess(failureOccurred).
			Render(views.ExtractRenderDeps(ctx))
	}

	plainText, hashedToken, err := a.tknManager.GenerateToken()
	if err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	resetPWToken := tokens.CreateResetPasswordToken(plainText, hashedToken)

	if err := a.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Hash:      resetPWToken.Hash,
		ExpiresAt: database.ConvertToPGTimestamptz(resetPWToken.GetExpirationTime()),
		Scope:     resetPWToken.GetScope(),
		UserID:    user.ID,
	}); err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	// TODO fix this error flow
	pwResetMail := &templates.PasswordResetMail{
		ResetPasswordLink: fmt.Sprintf(
			"%s://%s/reset-password?token=%s",
			a.cfg.App.AppScheme,
			a.cfg.App.AppHost,
			resetPWToken.GetPlainText(),
		),
	}

	textVersion, err := pwResetMail.GenerateTextVersion()
	if err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}
	htmlVersion, err := pwResetMail.GenerateHtmlVersion()
	if err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	_, err = a.queueClient.Insert(ctx.Request().Context(), queue.EmailJobArgs{
		To:          user.Mail,
		From:        a.cfg.App.DefaultSenderSignature,
		Subject:     "Password Reset Request",
		TextVersion: textVersion,
		HtmlVersion: htmlVersion,
	}, nil)
	if err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.ForgottenPasswordSuccess(false).Render(views.ExtractRenderDeps(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

func (a *Authentication) CreateResetPassword(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return a.InternalError(ctx)
	}

	return authentication.ResetPasswordPage(authentication.ResetPasswordPageProps{
		ResetToken: passwordResetToken.Token,
		CsrfToken:  csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

func (a *Authentication) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	hashedToken, err := a.tknManager.Hash(payload.Token)
	if err != nil {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	token, err := a.db.QueryTokenByHash(ctx.Request().Context(), hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
				HasError: true,
				Msg:      "The token is invalid. Please request a new one.",
			}).Render(views.ExtractRenderDeps(ctx))
		}

		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	if database.ConvertFromPGTimestamptzToTime(token.ExpiresAt).Before(time.Now()) &&
		token.Scope != tokens.ScopeResetPassword {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "The token has expired. Please request a new one.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	err = a.userModel.ChangePassword(ctx.Request().Context(),
		models.ChangeUserPasswordData{
			ID:              token.UserID,
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
			ResetToken: token.Hash,
		}

		for _, validationError := range valiErrs {
			switch validationError.GetFieldName() {
			case "Password":
				props.Errors[authentication.PasswordNotValid] = validationError.GetHumanExplanations()[0]
			case "ConfirmPassword":
				props.Errors[authentication.PasswordNotMatchConfirm] = validationError.GetHumanExplanations()[0]
			}
		}

		return authentication.ResetPasswordForm(props).Render(views.ExtractRenderDeps(ctx))
	}
	if err != nil {
		return a.InternalError(ctx)
	}

	if err := a.db.DeleteToken(ctx.Request().Context(), token.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return a.InternalError(ctx)
	}

	return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
		HasError: false,
	}).Render(views.ExtractRenderDeps(ctx))
}
