package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/templates"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

type SignupPageData struct {
	PreviousNameInput  string
	PreviousEmailInput string
	Errors             validator.ValidationErrors
	RenderPartial      bool
}

func SignupPage(ctx echo.Context, data SignupPageData) error {
	templateData := templates.RegisterUserData{CsrfToken: csrf.Token(ctx.Request())}

	if len(data.Errors) == 0 && data.RenderPartial {
		return templates.RegisterUserForm(templateData).Render(extractRenderDeps(ctx))
	}

	if len(data.Errors) > 0 {
		templateData.NameInput.Value = data.PreviousNameInput
		templateData.EmailInput.Value = data.PreviousEmailInput

		for _, validationError := range data.Errors {
			switch validationError.StructField() {
			case "Name":
				templateData.NameInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Mail":
				templateData.EmailInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
					Value:      validationError.Value().(string),
				}
			case "Password":
				templateData.PasswordInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "ConfirmPassword":
				templateData.ConfirmPassword = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: validationError.Param(),
				}
			case "MailRegistered":
				templateData.EmailInput = templates.TextInputData{
					Invalid:    true,
					InvalidMsg: "Email already registered",
				}
			}
		}

		return templates.RegisterUserForm(templateData).Render(extractRenderDeps(ctx))
	}

	return layouts.Base(templates.UserRegistrationWrapper(templateData)).Render(extractRenderDeps(ctx))
}

func SignupResponse(ctx echo.Context) error {
	return templates.UserRegisteredResponse().Render(extractRenderDeps(ctx))
}

// func RegisterUserForm(ctx echo.Context, data RegisterUserData) error {
// 	return ctx.Render(http.StatusOK, "user/__register_form", RenderOpts{
// 		Data: data,
// 	})
// }

// func RegisteredUser(ctx echo.Context) error {
// 	return ctx.Render(http.StatusOK, "user/__registered", RenderOpts{
// 		Data: nil,
// 	})
// }

// func PasswordForgotForm(ctx echo.Context) error {
// 	return ctx.Render(http.StatusOK, "user/forgot_password", RenderOpts{
// 		Layout: BaseLayout,
// 		Data: Csrf{
// 			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
// 		},
// 	})
// }

// func SendPasswordResetMail(ctx echo.Context) error {
// 	return ctx.Render(http.StatusOK, "user/__reset_email_send", RenderOpts{
// 		Data: nil,
// 	})
// }

// type ResetPasswordData struct {
// 	TokenInvalid    bool
// 	Token           string
// 	PasswordInput   InputData
// 	ConfirmPassword InputData
// 	CsrfField       template.HTML
// }

// func ResetPasswordForm(ctx echo.Context, data ResetPasswordData) error {
// 	return ctx.Render(http.StatusOK, "user/reset_password", RenderOpts{
// 		Layout: BaseLayout,
// 		Data:   data,
// 	})
// }

// // func ResetPassword(ctx echo.Context, data ResetPasswordData) error {
// // 	return ctx.Render(http.StatusOK, "user/reset_password", RenderOpts{
// // 		Layout: BaseLayout,
// // 		Data:   data,
// // 	})
// // }

// func ResetPasswordResponse(ctx echo.Context) error {
// 	return ctx.Render(http.StatusOK, "user/__reset_password_response", RenderOpts{
// 		Data: nil,
// 	})
// }
