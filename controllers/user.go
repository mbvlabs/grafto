package controllers

import (
	"html/template"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/MBvisti/grafto/views"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
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

	_, err := services.NewUser(ctx.Request().Context(), entity.NewUser{
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

	return c.views.RegisteredUser(ctx)
}
