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
	Path:        "/registrations/new",
	Method:      http.MethodGet,
	HandlerName: "New",
}

var StoreUser = Route{
	Name:        registrationNamePrefix + ".store_user",
	Path:        "/registrations",
	Method:      http.MethodPost,
	HandlerName: "Create",
}

var VerifyEmail = Route{
	Name:        registrationNamePrefix + ".verify_email",
	Path:        "/verify-email",
	Method:      http.MethodPost,
	HandlerName: "Update",
}
