package controllers

import (
	"database/sql"
	"errors"
	"html/template"
	"time"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/pkg/tokens"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

// CreateUser method    shows the form to create the user
func (c *Controller) CreateUser(ctx echo.Context) error {
	return c.views.RegisterUser(ctx)
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
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

		return c.InternalError(ctx)
	}

	user, err := services.NewUser(ctx.Request().Context(), entity.NewUser{
		Name:            payload.UserName,
		Mail:            payload.Mail,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	}, &c.db, c.validate)
	if err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			telemetry.Logger.Info("internal error", "ok", ok)
		}

		if len(e) == 0 {
			telemetry.Logger.WarnContext(ctx.Request().Context(), "an unrecoverable error occurred", "error", err)

			ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
			ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

			return c.InternalError(ctx)
		}

		viewData := views.RegisterUserData{
			NameInput: views.InputData{
				OldValue: payload.UserName,
			},
			EmailInput: views.InputData{
				OldValue: payload.Mail,
			},
			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
		}

		for _, validationError := range e {
			switch validationError.StructField() {
			case "Name":
				viewData.NameInput = views.InputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					OldValue:   validationError.Value(),
				}
			case "Mail":
				viewData.EmailInput = views.InputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					OldValue:   validationError.Value(),
				}
			case "Password":
				viewData.PasswordInput = views.InputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				viewData.ConfirmPassword = views.InputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "MailRegistered":
				viewData.EmailInput = views.InputData{
					Invalid:    true,
					InvalidMsg: "Email already registered",
				}
			}
		}

		return c.views.RegisterUserForm(ctx, viewData)
	}

	tkn, err := tokens.CreateActivationToken()
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	telemetry.Logger.Info("controller hashed", "tkn", tkn.HashedToken, "raw", tkn.GetRawToken())

	if err := c.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		Hash:      tkn.HashedToken,
		ExpiresAt: tkn.GetExpirationTime(),
		Scope:     tkn.GetScope(),
		UserID:    user.ID,
	}); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	if err := c.mail.Send(ctx.Request().Context(),
		user.Mail, "newsletter@mortenvistisen.com", "Testing Email Confirmation", "confirm_email",
		mail.ConfirmPassword{
			Token: tkn.GetRawToken(),
		}); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return c.views.RegisteredUser(ctx)
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

	hashedToken, err := tokens.HashToken(tkn.Token)
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	// telemetry.Logger.Info("hashed token", "hashed_token", hashedToken)
	token, err := c.db.GetTokenByHash(ctx.Request().Context(), hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.views.EmailValidation(ctx, views.EmailValidationData{
				TokenInvalid: true,
			})
		}

		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	if token.ExpiresAt.Before(time.Now()) && token.Scope != tokens.ScopeEmailVerification {
		return c.views.EmailValidation(ctx, views.EmailValidationData{
			TokenInvalid: true,
		})
	}

	confirmTime := time.Now()
	user, err := c.db.ConfirmUserEmail(ctx.Request().Context(), database.ConfirmUserEmailParams{
		ID:             token.UserID,
		UpdatedAt:      confirmTime,
		MailVerifiedAt: sql.NullTime{Time: confirmTime, Valid: true},
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

	if err := services.CreateAuthenticatedSession(ctx.Request(), ctx.Response(), user.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return c.views.EmailValidation(ctx, views.EmailValidationData{
		TokenInvalid: false,
	})
}

type PasswordResetRequestPayload struct {
	Email string `form:"email"`
}

func (c *Controller) RenderPasswordForgotForm(ctx echo.Context) error {
	return c.views.PasswordForgotForm(ctx)
}

// func (c *Controller) ResetPassword(ctx echo.Context) error {
// 	var tkn PasswordResetRequestPayload
// 	if err := ctx.Bind(&tkn); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

// 		return c.InternalError(ctx)
// 	}

// 	return c.views.EmailValidation(ctx)
// }
