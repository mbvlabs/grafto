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
	Path:        "/login",
	Method:      http.MethodGet,
	HandlerName: "CreateAuthenticatedSession",
}

var StoreAuthSession = Route{
	Name:        authNamePrefix + ".store_auth_session",
	Path:        "/login",
	Method:      http.MethodPost,
	HandlerName: "StoreAuthenticatedSession",
	Middleware: []string{
		"LoginRateLimiter",
	},
}

var DestroyAuthSession = Route{
	Name:        authNamePrefix + ".destroy_auth_session",
	Path:        "/logout",
	Method:      http.MethodGet,
	HandlerName: "DestroyAuthenticatedSession",
}

var ForgotPasswordPage = Route{
	Name:        authNamePrefix + ".forgot_password_page",
	Path:        "/forgot-password",
	Method:      http.MethodGet,
	HandlerName: "CreatePasswordReset",
}

var StoreForgotPassword = Route{
	Name:        authNamePrefix + ".store_forgot_password",
	Path:        "/forgot-password",
	Method:      http.MethodPost,
	HandlerName: "StorePasswordReset",
}

var ResetPasswordPage = Route{
	Name:        authNamePrefix + ".reset_password_page",
	Path:        "/reset-password",
	Method:      http.MethodGet,
	HandlerName: "CreateResetPassword",
}

var StoreResetPasswordPage = Route{
	Name:        authNamePrefix + ".store_reset_password_page",
	Path:        "/reset-password",
	Method:      http.MethodPost,
	HandlerName: "StoreResetPassword",
}
