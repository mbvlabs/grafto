package controllers

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/entity"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/repository/database"
	"github.com/mbv-labs/grafto/services"
	"github.com/mbv-labs/grafto/views"
	"github.com/mbv-labs/grafto/views/authentication"
)

// CreateUser method    shows the form to create the user
func (c *Controller) CreateUser(ctx echo.Context) error {
	return authentication.RegisterPage(authentication.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type StoreUserPayload struct {
	UserName        string `form:"user_name"`
	Mail            string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

// StoreUser method    stores the new user
func (c *Controller) StoreUser(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	user, err := services.NewUser(ctx.Request().Context(), entity.NewUser{
		Name:            payload.UserName,
		Mail:            payload.Mail,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	}, &c.db, c.validate, c.cfg.Auth.PasswordPepper)
	if err != nil {
		telemetry.Logger.Info("error", "err", err)
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			telemetry.Logger.WarnContext(ctx.Request().Context(), "an unrecoverable error occurred", "error", err)

			return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
		}

		if len(e) == 0 {
			telemetry.Logger.WarnContext(ctx.Request().Context(), "an unrecoverable error occurred", "error", err)

			return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
		}

		props := authentication.RegisterFormProps{
			NameInput: views.InputElementError{
				OldValue: payload.UserName,
			},
			EmailInput: views.InputElementError{
				OldValue: payload.Mail,
			},
			CsrfToken: csrf.Token(ctx.Request()),
		}

		for _, validationError := range e {
			switch validationError.StructField() {
			case "Name":
				props.NameInput.Invalid = true
				props.NameInput.InvalidMsg = validationError.Param()
			case "MailRegistered":
				props.EmailInput.Invalid = true
				props.EmailInput.InvalidMsg = validationError.Param()
			case "Password", "ConfirmPassword":
				props.PasswordInput = views.InputElementError{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
				props.ConfirmPassword = views.InputElementError{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			}
		}

		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	plainText, hashedToken, err := c.tknManager.GenerateToken()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	activationToken := tokens.CreateActivationToken(plainText, hashedToken)

	if err := c.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Hash:      activationToken.Hash,
		ExpiresAt: database.ConvertToPGTimestamptz(activationToken.GetExpirationTime()),
		Scope:     activationToken.GetScope(),
		UserID:    user.ID,
	}); err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}
	userSignupMail := templates.UserSignupWelcomeMail{
		ConfirmationLink: fmt.Sprintf(
			"%s://%s/verify-email?token=%s",
			c.cfg.App.AppScheme,
			c.cfg.App.AppHost,
			activationToken.GetPlainText(),
		),
		UnsubscribeLink: "", // TODO implement
	}
	textVersion, err := userSignupMail.GenerateTextVersion()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}
	htmlVersion, err := userSignupMail.GenerateHtmlVersion()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	_, err = c.queueClient.Insert(ctx.Request().Context(), queue.EmailJobArgs{
		To:          user.Mail,
		From:        c.cfg.App.DefaultSenderSignature,
		Subject:     "Thanks for signing up!",
		TextVersion: textVersion,
		HtmlVersion: htmlVersion,
	}, nil)
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.RegisterResponse("You're now registered", "You should receive an email soon to validate your account.", false).Render(views.ExtractRenderDeps(ctx))
}
