package paths

import (
	"context"

	"github.com/a-h/templ"
	"github.com/mbvlabs/grafto/views/contexts"
)

const (
	HomePage           = "homePage"
	AboutPage          = "aboutPage"
	LoginPage          = "loginPage"
	Login              = "login"
	ForgotPasswordPage = "forgotPasswordPage"
	ForgotPassword     = "forgotPassword"
	ResetPasswordPage  = "resetPasswordPage"
	ResetPassword      = "resetPassword"
	DashboardHomePage  = "dashbordHomePage"
	RegisterPage       = "registerPage"
	RegisterUser       = "registerUser"
	VerifyEmailPage    = "verifyEmailPage"
)

func Get(ctx context.Context, name string) string {
	return contexts.ExtractApp(ctx).Routes[name]
}

func GetSafeURL(ctx context.Context, name string) templ.SafeURL {
	return templ.SafeURL(contexts.ExtractApp(ctx).Routes[name])
}
