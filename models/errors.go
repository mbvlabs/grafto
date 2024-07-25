package models

import "errors"

var (
	ErrFailValidation    = errors.New("the object failed validations")
	ErrUserAlreadyExists = errors.New("an user with the provided email already exists")
)
