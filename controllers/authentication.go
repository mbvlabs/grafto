package controllers

import (
	"errors"
	"time"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/pkg/tokens"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/MBvisti/grafto/views/authentication"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

func (c *Controller) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type UserLoginPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (c *Controller) StoreAuthenticatedSession(ctx echo.Context) error {
	var payload UserLoginPayload
	if err := ctx.Bind(&payload); err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not parse UserLoginPayload", "error", err)

		return authentication.LoginResponse(true).Render(views.ExtractRenderDeps(ctx))
	}

	authenticatedUser, err := services.AuthenticateUser(
		ctx.Request().Context(), services.AuthenticateUserPayload{
			Email:    payload.Mail,
			Password: payload.Password,
		}, &c.db, c.cfg.Auth.PasswordPepper)
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not authenticate user", "error", err)

		errMsg := "An error occurred while trying to authenticate you. Please try again."

		switch err {
		case services.ErrPasswordNotMatch, services.ErrUserNotExist:
			errMsg = "The password you entered is incorrect."
		case services.ErrEmailNotValidated:
			errMsg = "You need to verify your email before you can log in. Please check your inbox for a verification email."
		}
		return authentication.LoginForm(csrf.Token(ctx.Request()), authentication.LoginFormProps{
			HasError: true,
			ErrMsg:   errMsg,
		}).Render(views.ExtractRenderDeps(ctx))
	}

	session, err := c.authSessionStore.Get(ctx.Request(), "ua")
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}
	if err := services.CreateAuthenticatedSession(*session, authenticatedUser.ID, c.cfg); err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return authentication.LoginResponse(true).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.LoginResponse(false).Render(views.ExtractRenderDeps(ctx))
}

func (c *Controller) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(authentication.ForgottenPasswordPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type StorePasswordResetPayload struct {
	Mail string `form:"email"`
}

func (c *Controller) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	user, err := c.db.QueryUserByMail(ctx.Request().Context(), payload.Mail)
	if err != nil {
		failureOccurred := true
		if errors.Is(err, pgx.ErrNoRows) {
			failureOccurred = false
		}

		return authentication.ForgottenPasswordSuccess(failureOccurred).Render(views.ExtractRenderDeps(ctx))
	}

	plainText, hashedToken, err := c.tknManager.GenerateToken()
	if err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	resetPWToken := tokens.CreateResetPasswordToken(plainText, hashedToken)

	if err := c.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Hash:      resetPWToken.Hash,
		ExpiresAt: database.ConvertToPGTimestamptz(resetPWToken.GetExpirationTime()),
		Scope:     resetPWToken.GetScope(),
		UserID:    user.ID,
	}); err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	if err := c.mail.Send(ctx.Request().Context(),
		user.Mail, c.cfg.App.DefaultSenderSignature, "Password Reset Request", "password_reset",
		mail.ConfirmPassword{
			Token: resetPWToken.GetPlainText(),
		}); err != nil {
		return authentication.ForgottenPasswordSuccess(true).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.ForgottenPasswordSuccess(false).Render(views.ExtractRenderDeps(ctx))
}

type PasswordResetToken struct {
	Token string `query:"token"`
}

func (c *Controller) CreateResetPassword(ctx echo.Context) error {
	var passwordResetToken PasswordResetToken
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return c.InternalError(ctx)
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

func (c *Controller) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	hashedToken, err := c.tknManager.Hash(payload.Token)
	if err != nil {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	token, err := c.db.QueryTokenByHash(ctx.Request().Context(), hashedToken)
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

	if database.ConvertFromPGTimestamptzToTime(token.ExpiresAt).Before(time.Now()) && token.Scope != tokens.ScopeResetPassword {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "The token has expired. Please request a new one.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	user, err := c.db.QueryUser(ctx.Request().Context(), token.UserID)
	if err != nil {
		return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
			HasError: true,
			Msg:      "An error occurred while trying to reset your password. Please try again.",
		}).Render(views.ExtractRenderDeps(ctx))
	}

	_, err = services.UpdateUser(ctx.Request().Context(), entity.UpdateUser{
		Name:            user.Name,
		Mail:            user.Mail,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
		ID:              user.ID,
	}, &c.db, c.validate, c.cfg.Auth.PasswordPepper)
	if err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			telemetry.Logger.Info("internal error", "ok", ok)
		}

		if len(e) == 0 {
			return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
				HasError: true,
				Msg:      "An error occurred while trying to reset your password. Please try again.",
			}).Render(views.ExtractRenderDeps(ctx))
		}

		props := authentication.ResetPasswordFormProps{
			CsrfToken:  csrf.Token(ctx.Request()),
			ResetToken: token.Hash,
		}

		for _, validationError := range e {
			switch validationError.StructField() {
			case "Password", "ConfirmPassword":
				props.Password = views.InputElementError{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
				props.ConfirmPassword = views.InputElementError{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			}
		}

		return authentication.ResetPasswordForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	if err := c.db.DeleteToken(ctx.Request().Context(), token.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return authentication.ResetPasswordResponse(authentication.ResetPasswordResponseProps{
		HasError: false,
	}).Render(views.ExtractRenderDeps(ctx))
}

type VerifyEmail struct {
	Token string `query:"token"`
}

// VerifyEmail method    verifies the email the user provided during signup
func (c *Controller) VerifyEmail(ctx echo.Context) error {
	var tkn VerifyEmail
	if err := ctx.Bind(&tkn); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

		return c.InternalError(ctx)
	}

	hashedToken, err := c.tknManager.Hash(tkn.Token)
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	token, err := c.db.QueryTokenByHash(ctx.Request().Context(), hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authentication.VerifyEmailPage(true, views.Head{}).Render(views.ExtractRenderDeps(ctx))
		}

		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	if database.ConvertFromPGTimestamptzToTime(token.ExpiresAt).Before(time.Now()) && token.Scope != tokens.ScopeEmailVerification {
		return authentication.VerifyEmailPage(true, views.Head{}).Render(views.ExtractRenderDeps(ctx))
	}

	confirmTime := time.Now()
	user, err := c.db.ConfirmUserEmail(ctx.Request().Context(), database.ConfirmUserEmailParams{
		ID:             token.UserID,
		UpdatedAt:      database.ConvertToPGTimestamptz(confirmTime),
		MailVerifiedAt: database.ConvertToPGTimestamptz(confirmTime),
	})
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	if err := c.db.DeleteToken(ctx.Request().Context(), token.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	session, err := c.authSessionStore.Get(ctx.Request(), "ua")
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	authSession := services.CreateAuthenticatedSession(*session, user.ID, c.cfg)
	if err := authSession.Save(ctx.Request(), ctx.Response()); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return authentication.VerifyEmailPage(false, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}
