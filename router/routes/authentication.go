package routes

import (
	"net/http"
)

const authNamePrefix = "auth"

var Authentication = []Route{
	LoginPage,
	StoreAuthSession,
	DestroyAuthSession,
	ForgotPasswordPage,
	StoreForgotPassword,
	ResetPasswordPage,
	StoreResetPasswordPage,
}

var LoginPage = Route{
	Name:        authNamePrefix + ".login_page",
	Path:        "/sessions/new",
	Method:      http.MethodGet,
	HandlerName: "New",
}

var StoreAuthSession = Route{
	Name:        authNamePrefix + ".store_auth_session",
	Path:        "/sessions",
	Method:      http.MethodPost,
	HandlerName: "Create",
	Middleware: []string{
		"LoginRateLimiter",
	},
}

var DestroyAuthSession = Route{
	Name:        authNamePrefix + ".destroy_auth_session",
	Path:        "/sessions",
	Method:      http.MethodDelete,
	HandlerName: "Destroy",
}

var ForgotPasswordPage = Route{
	Name:        authNamePrefix + ".forgot_password_page",
	Path:        "/password-resets/new",
	Method:      http.MethodGet,
	HandlerName: "NewPasswordReset",
}

var StoreForgotPassword = Route{
	Name:        authNamePrefix + ".store_forgot_password",
	Path:        "/password-resets",
	Method:      http.MethodPost,
	HandlerName: "CreatePasswordReset",
}

var ResetPasswordPage = Route{
	Name:        authNamePrefix + ".reset_password_page",
	Path:        "/password-resets/edit",
	Method:      http.MethodGet,
	HandlerName: "EditPasswordReset",
}

var StoreResetPasswordPage = Route{
	Name:        authNamePrefix + ".store_reset_password_page",
	Path:        "/password-resets",
	Method:      http.MethodPut,
	HandlerName: "UpdatePasswordReset",
}
