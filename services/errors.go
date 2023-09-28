package services

import "github.com/pkg/errors"

var (
	ErrInvalidInput      = errors.New("one or more of the provided inputs are not valid")
	ErrEmailNotValidated = errors.New("user email not validated")
	ErrUserNotExist      = errors.New("user have not been registered")
	ErrPasswordNotMatch  = errors.New("provided password does not match our records")
)
