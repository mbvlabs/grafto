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
	Name:         authNamePrefix + ".login_page",
	Path:         "/sessions/new",
	Method:       http.MethodGet,
	Handler:      "Sessions",
	HandleMethod: "New",
}

var StoreAuthSession = Route{
	Name:         authNamePrefix + ".store_auth_session",
	Path:         "/sessions",
	Method:       http.MethodPost,
	Handler:      "Sessions",
	HandleMethod: "Create",
	Middleware: []string{
		"LoginRateLimiter",
	},
}

var DestroyAuthSession = Route{
	Name:         authNamePrefix + ".destroy_auth_session",
	Path:         "/sessions",
	Method:       http.MethodDelete,
	Handler:      "Sessions",
	HandleMethod: "Destroy",
}

var ForgotPasswordPage = Route{
	Name:         authNamePrefix + ".forgot_password_page",
	Path:         "/password-resets/new",
	Method:       http.MethodGet,
	Handler:      "Sessions",
	HandleMethod: "NewPasswordReset",
}

var StoreForgotPassword = Route{
	Name:         authNamePrefix + ".store_forgot_password",
	Path:         "/password-resets",
	Method:       http.MethodPost,
	Handler:      "Sessions",
	HandleMethod: "CreatePasswordReset",
}

var ResetPasswordPage = Route{
	Name:         authNamePrefix + ".reset_password_page",
	Path:         "/password-resets/edit",
	Method:       http.MethodGet,
	Handler:      "Sessions",
	HandleMethod: "EditPasswordReset",
}

var StoreResetPasswordPage = Route{
	Name:         authNamePrefix + ".store_reset_password_page",
	Path:         "/password-resets",
	Method:       http.MethodPut,
	Handler:      "Sessions",
	HandleMethod: "UpdatePasswordReset",
}
