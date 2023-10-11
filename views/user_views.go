package views

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

type Csrf struct {
	CsrfField template.HTML
}

type RegisterUserData struct {
	NameInput       InputData
	EmailInput      InputData
	PasswordInput   InputData
	ConfirmPassword InputData
	CsrfField       template.HTML
}

func (v Views) RegisterUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/register", RenderOpts{
		Layout: BaseLayout,
		Data: RegisterUserData{
			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
		},
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

func (v Views) PasswordForgotForm(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/forgot_password", RenderOpts{
		Layout: BaseLayout,
		Data: Csrf{
			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
		},
	})
}
