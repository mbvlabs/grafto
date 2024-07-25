package services

import "github.com/pkg/errors"

var (
	ErrEmailNotValidated = errors.New("user email not validated")
	ErrUserNotExist      = errors.New("user have not been registered")
	ErrPasswordNotMatch  = errors.New("provided password does not match our records")
)
