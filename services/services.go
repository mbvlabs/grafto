package services

import (
	"github.com/MBvisti/grafto/repository/database"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
)

type Services struct {
	authSessionStore *sessions.CookieStore
	db               database.Queries
	validator        *validator.Validate
}

func NewServices(authSessionStore *sessions.CookieStore, db database.Queries, v *validator.Validate) Services {
	validator := registerStructValidations(v)

	return Services{
		authSessionStore,
		db,
		validator,
	}
}
