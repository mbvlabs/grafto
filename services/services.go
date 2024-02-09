package services

import "github.com/gorilla/sessions"

type Services struct {
	authSessionStore *sessions.CookieStore
}

func NewServices(authSessionStore *sessions.CookieStore) Services {
	return Services{
		authSessionStore,
	}
}
