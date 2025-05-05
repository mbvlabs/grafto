package paths

var (
	CreateAuthenticatedSession = register(
		Path{Name: "auth.login", URL: "/login"},
	)
	StoreAuthenticatedSession = register(
		Path{Name: "auth.store_authenticated_session", URL: "/login"},
	)
	DestroyAuthenticatedSession = register(
		Path{Name: "auth.destroy_authenticated_session", URL: "/logout"},
	)
	CreateForgotPassword = register(
		Path{Name: "auth.create_forgot_password", URL: "/forgot-password"},
	)
	StoreForgotPassword = register(
		Path{Name: "auth.store_forgot_password", URL: "/forgot-password"},
	)
	CreateResetPassword = register(
		Path{Name: "auth.create_reset_password", URL: "/reset-password"},
	)
	StoreResetPassword = register(
		Path{Name: "auth.store_reset_password", URL: "/reset-password"},
	)
)
