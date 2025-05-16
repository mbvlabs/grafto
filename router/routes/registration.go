package routes

import "net/http"

const (
	registrationNamePrefix = "registration"
)

var Registration = []Route{
	CreateUserPage,
	StoreUser,
	VerifyEmail,
}

var CreateUserPage = Route{
	Name:        registrationNamePrefix + ".create_user",
	Path:        "/register",
	Method:      http.MethodGet,
	HandlerName: "CreateUser",
}

var StoreUser = Route{
	Name:        registrationNamePrefix + ".store_user",
	Path:        "/register",
	Method:      http.MethodPost,
	HandlerName: "StoreUser",
}

var VerifyEmail = Route{
	Name:        registrationNamePrefix + ".verify_email",
	Path:        "/verify-email",
	Method:      http.MethodPost,
	HandlerName: "VerifyUserEmail",
}
