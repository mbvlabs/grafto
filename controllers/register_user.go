package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CreateUser method    shows the form to create the user
func (c *Controller) CreateUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/register", views.RenderOpts{
		Data: nil,
	})
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
		panic("yooyoyyoy")
	}

	_, err := services.NewUser(ctx.Request().Context(), services.NewUserData{
		Name:            payload.UserName,
		Mail:            payload.Mail,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	}, &c.db)
	if err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			telemetry.Logger.Info("internal error", "ok", ok)
		}
		telemetry.Logger.Info("error payload", "error", e)
		return err
	}

	return ctx.Render(http.StatusOK, "user/__registered", views.RenderOpts{
		Data: nil,
	})
}

func (c *Controller) AuthenticateUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/sign_in", views.RenderOpts{
		Data: nil,
	})
}
