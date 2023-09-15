package views

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type RegisterUserData struct {
	NameInput       InputData
	EmailInput      InputData
	PasswordInput   InputData
	ConfirmPassword InputData
}

func (v Views) RegisterUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/register", RenderOpts{
		Layout: BaseLayout,
	})
}

func (v Views) RegisterUserForm(ctx echo.Context, data RegisterUserData) error {
	return ctx.Render(http.StatusOK, "user/__register_form", RenderOpts{
		Data: data,
	})
}

func (v Views) RegisteredUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/__registered", RenderOpts{
		Data: nil,
	})
}
