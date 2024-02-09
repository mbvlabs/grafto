package controllers

import (
	"time"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/pkg/tokens"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/MBvisti/grafto/views/authentication"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
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

	emailJob, err := queue.NewEmailJob(
		user.Mail, "info@mortenvistisen.com", "please confirm your email", "confirm_email", mail.ConfirmPassword{
			Token: activationToken.GetPlainText(),
		})
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	if err := c.queue.Push(ctx.Request().Context(), emailJob); err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.RegisterResponse("You're now registered", "You should receive an email soon to validate your account.", false).Render(views.ExtractRenderDeps(ctx))
}

// func (c *Controller) RenderPasswordForgotForm(ctx echo.Context) error {
// 	return c.views.PasswordForgotForm(ctx)
// }

// type PasswordResetRequestPayload struct {
// 	Email string `form:"email"`
// }

// func (c *Controller) SendPasswordResetEmail(ctx echo.Context) error {
// 	var payload PasswordResetRequestPayload
// 	if err := ctx.Bind(&payload); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create") // TODO:

// 		return c.InternalError(ctx)
// 	}

// 	user, err := c.db.QueryUserByMail(ctx.Request().Context(), payload.Email)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return c.views.SendPasswordResetMail(ctx)
// 		}

// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create") // TODO:

// 		return c.InternalError(ctx)
// 	}

// 	plainText, hashedToken, err := c.tknManager.GenerateToken()
// 	if err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login") // TODO:

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	resetPWToken := tokens.CreateResetPasswordToken(plainText, hashedToken)

// 	if err := c.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
// 		ID:        uuid.New(),
// 		CreatedAt: time.Now(),
// 		Hash:      resetPWToken.Hash,
// 		ExpiresAt: resetPWToken.GetExpirationTime(),
// 		Scope:     resetPWToken.GetScope(),
// 		UserID:    user.ID,
// 	}); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	if err := c.mail.Send(ctx.Request().Context(),
// 		user.Mail, "newsletter@mortenvistisen.com", "Password Reset Request", "password_reset",
// 		mail.ConfirmPassword{
// 			Token: resetPWToken.GetPlainText(),
// 		}); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	return c.views.SendPasswordResetMail(ctx)
// }

// type ResetPasswordPayload struct {
// 	Password        string `form:"password"`
// 	ConfirmPassword string `form:"confirm_password"`
// 	Token           string `form:"token"`
// }

// func (c *Controller) ResetPasswordForm(ctx echo.Context) error {
// 	var verifyToken VerifyEmail
// 	if err := ctx.Bind(&verifyToken); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

// 		return c.InternalError(ctx)
// 	}

// 	telemetry.Logger.Info("this is the token", "tkn", verifyToken.Token)

// 	return c.views.ResetPasswordForm(ctx, views.ResetPasswordData{
// 		Token:        verifyToken.Token,
// 		TokenInvalid: false,
// 		CsrfField:    template.HTML(csrf.TemplateField(ctx.Request())),
// 	})
// }

// func (c *Controller) ResetPassword(ctx echo.Context) error {
// 	var payload ResetPasswordPayload
// 	if err := ctx.Bind(&payload); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

// 		return c.InternalError(ctx)
// 	}

// 	hashedToken, err := c.tknManager.Hash(payload.Token)
// 	if err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	token, err := c.db.QueryTokenByHash(ctx.Request().Context(), hashedToken)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			telemetry.Logger.Error("token invalid because it was not found")
// 			return c.views.ResetPasswordForm(ctx, views.ResetPasswordData{
// 				TokenInvalid: true,
// 			})
// 		}

// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	telemetry.Logger.Info("this is token", "user_id", token.UserID)

// 	if token.ExpiresAt.Before(time.Now()) && token.Scope != tokens.ScopeResetPassword {
// 		telemetry.Logger.Error("token invalid because time or scope issue")
// 		return c.views.ResetPasswordForm(ctx, views.ResetPasswordData{
// 			TokenInvalid: true,
// 		})
// 	}

// 	user, err := c.db.QueryUser(ctx.Request().Context(), token.UserID)
// 	if err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	_, err = services.UpdateUser(ctx.Request().Context(), entity.UpdateUser{
// 		Name:            user.Name,
// 		Mail:            user.Mail,
// 		Password:        payload.Password,
// 		ConfirmPassword: payload.ConfirmPassword,
// 		ID:              user.ID,
// 	}, &c.db, c.validate)
// 	if err != nil {
// 		e, ok := err.(validator.ValidationErrors)
// 		if !ok {
// 			telemetry.Logger.Info("internal error", "ok", ok)
// 		}

// 		if len(e) == 0 {
// 			telemetry.Logger.WarnContext(ctx.Request().Context(), "an unrecoverable error occurred", "error", err)

// 			ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 			ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

// 			return c.InternalError(ctx)
// 		}

// 		viewData := views.ResetPasswordData{
// 			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
// 		}

// 		for _, validationError := range e {
// 			switch validationError.StructField() {
// 			case "Password":
// 				viewData.PasswordInput = views.InputData{
// 					Invalid:    true,
// 					InvalidMsg: validationError.Param(),
// 				}
// 			case "ConfirmPassword":
// 				viewData.ConfirmPassword = views.InputData{
// 					Invalid:    true,
// 					InvalidMsg: validationError.Param(),
// 				}
// 			}
// 		}

// 		return c.views.ResetPasswordForm(ctx, viewData)
// 	}

// 	if err := c.db.DeleteToken(ctx.Request().Context(), token.ID); err != nil {
// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
// 		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

// 		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
// 		return c.InternalError(ctx)
// 	}

// 	return c.views.ResetPasswordResponse(ctx)
// }
