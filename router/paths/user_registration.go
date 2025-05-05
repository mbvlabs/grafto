package paths

var (
	CreateUser = register(
		Path{Name: "user_registration.create_user", URL: "/register"},
	)
	StoreUser = register(
		Path{Name: "user_registration.store_user", URL: "/register"},
	)
	VerifyEmail = register(
		Path{Name: "user_registration.verify_email", URL: "/verify-email"},
	)
)
