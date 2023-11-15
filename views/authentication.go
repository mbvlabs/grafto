package views

import "github.com/labstack/echo/v4"

type LoginPageData struct {
	CouldNotAuthenticate bool
	EmailNotVerified     bool
	RenderPartial        bool
}

func LoginPage(ctx echo.Context, data LoginPageData) error {
	//	return ctx.Render(http.StatusOK, "user/login", RenderOpts{
	//		Layout: BaseLayout,
	//		Data: LoginForm{
	//			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
	//		},
	//	})
	return nil
}

func LoginResponse(ctx echo.Context) error {
	//	return ctx.Render(http.StatusOK, "user/login", RenderOpts{
	//		Layout: BaseLayout,
	//		Data: LoginForm{
	//			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
	//		},
	//	})
	return nil
}

// type LoginForm struct {
// 	EmailNeedsVerification bool
// 	CouldNotAuthenticate   bool
// 	CsrfField              template.HTML
// }

// func LoginForm(ctx echo.Context, data LoginForm) error {
// 	return ctx.Render(http.StatusOK, "user/__login_form", RenderOpts{
// 		Data: data,
// 	})
// }

// func Authenticated(ctx echo.Context) error {
// 	return ctx.Render(http.StatusOK, "user/__authenticated", RenderOpts{
// 		Data: nil,
// 	})
// }

// type EmailValidationData struct {
// 	TokenInvalid bool
// }

// func EmailValidation(ctx echo.Context, data EmailValidationData) error {
// 	return ctx.Render(http.StatusOK, "user/email_validation", RenderOpts{
// 		Layout: BaseLayout,
// 		Data:   data,
// 	})
// }
