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
	Name:         registrationNamePrefix + ".create_user",
	Path:         "/registrations/new",
	Method:       http.MethodGet,
	Handler:      "Registrations",
	HandleMethod: "New",
}

var StoreUser = Route{
	Name:         registrationNamePrefix + ".store_user",
	Path:         "/registrations",
	Method:       http.MethodPost,
	Handler:      "Registrations",
	HandleMethod: "Create",
}

var VerifyEmail = Route{
	Name:         registrationNamePrefix + ".verify_email",
	Path:         "/verify-email",
	Method:       http.MethodPost,
	Handler:      "Registrations",
	HandleMethod: "Update",
}
