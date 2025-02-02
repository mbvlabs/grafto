package paths

import (
	"context"

	"github.com/a-h/templ"
)

type Route string

func (r Route) Name() string {
	return string(r)
}

const (
	HomePage               Route = "homePage"
	AboutPage              Route = "aboutPage"
	LoginPage              Route = "loginPage"
	Login                  Route = "login"
	ForgotPasswordPage     Route = "forgotPasswordPage"
	ForgotPassword         Route = "forgotPassword"
	ResetPasswordPage      Route = "resetPasswordPage"
	ResetPassword          Route = "resetPassword"
	RegisterPage           Route = "registerPage"
	RegisterUser           Route = "registerUser"
	VerifyEmailPage        Route = "verifyEmailPage"
	ArticlePage            Route = "articlePage"
	ArticlesPage           Route = "articlesPage"
	ProjectsPage           Route = "projectsPage"
	NewslettersPage        Route = "newslettersPage"
	SubscribeEvent         Route = "subscribeEvent"
	VerifySubEvent         Route = "verifySubEvent"
	UnsubscribeEvent       Route = "unsubscribeEvent"
	DashboardHomePage      Route = "dashbordHomePage"
	DashboardNewsletter    Route = "dashboardNewsletter"
	DashboardNewsletterNew Route = "dashboardNewsletterNew"
)

func Get(ctx context.Context, route Route) string {
	return ctx.Value(route).(string)
	// return contexts.ExtractApp(ctx).Routes[string(route)]
}

func GetSafeURL(ctx context.Context, route Route) templ.SafeURL {
	return templ.SafeURL(Get(ctx, route))
}
