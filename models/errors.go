package models

import "errors"

var (
	ErrDomainValidation = errors.New("the provided payload failed validations")
	ErrMustBeAdmin      = errors.New(
		"this action requires the actor to be an admin",
	)
)
