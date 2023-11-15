package controllers

import (
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) CreateAuthenticatedSession(ctx echo.Context) error {
	shouldSwap := false
	if ctx.QueryParam("should_swap") == "true" {
		shouldSwap = true
	}

	return views.LoginPage(ctx, views.LoginPageData{
		RenderPartial: shouldSwap,
	})
}

type UserLoginPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (c *Controller) StoreAuthenticatedSession(ctx echo.Context) error {
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
		responseData := views.LoginPageData{
			RenderPartial: true,
		}

		switch err {
		case services.ErrPasswordNotMatch:
			responseData.CouldNotAuthenticate = true
		case services.ErrUserNotExist:
			responseData.CouldNotAuthenticate = true
		case services.ErrEmailNotValidated:
			responseData.EmailNotVerified = true
		default:
			return err
		}
		return views.LoginPage(ctx, responseData)
	}

	if err := services.CreateAuthenticatedSession(ctx.Request(), ctx.Response(), authenticatedUser.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return c.InternalError(ctx)
	}

	return views.LoginResponse(ctx)
}
