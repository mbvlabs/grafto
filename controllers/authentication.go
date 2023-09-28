package controllers

import (
	"html/template"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
)

func (c *Controller) Login(ctx echo.Context) error {
	return c.views.LoginPage(ctx)
}

type UserLoginPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (c *Controller) Authenticate(ctx echo.Context) error {
	var payload UserLoginPayload
	if err := ctx.Bind(&payload); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	authenticatedUser, err := services.AuthenticateUser(
		ctx.Request().Context(), services.AuthenticateUserPayload{
			Email:    payload.Mail,
			Password: payload.Password,
		}, &c.db)
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		responseData := views.LoginForm{
			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
		}
		switch err {
		case services.ErrPasswordNotMatch:
			responseData.CouldNotAuthenticate = true
			return c.views.LoginForm(ctx, responseData)
		case services.ErrUserNotExist:
			responseData.CouldNotAuthenticate = true
			return c.views.LoginForm(ctx, responseData)
		case services.ErrEmailNotValidated:
			responseData.EmailNeedsVerification = true
			return c.views.LoginForm(ctx, responseData)
		default:
			return err
		}
	}

	if err := services.CreateAuthenticatedSession(ctx.Request(), ctx.Response(), authenticatedUser.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return c.views.Authenticated(ctx)
}
