package routes

import "github.com/labstack/echo/v4"

type Route struct {
	Name         string
	Path         string
	Handler      string
	HandleMethod string
	Method       string
	Middleware   []func(next echo.HandlerFunc) echo.HandlerFunc
}

var BuildRoutes = func() []Route {
	var r []Route

	r = append(
		r,
		LandingPage,
		AboutPage,
		Redirect.Route,
		Health,
		Robots,
		Sitemap,
		CssEntrypoint,
		AllCss,
		JsEntrypoint,
		AllJs,
		Favicon16,
		Favicon32,
		DashboardHome,
		DashboardUsers,
		DashboardUserEdit,
		DashboardUserUpdate,
		DashboardUserDelete,
		DashboardUserToggleAdmin,
		CreateUserPage,
		StoreUser,
		VerifyEmail,
		LoginPage,
		StoreAuthSession,
		DestroyAuthSession,
		ForgotPasswordPage,
		StoreForgotPassword,
		ResetPasswordPage,
		StoreResetPasswordPage,
	)

	return r
}()
