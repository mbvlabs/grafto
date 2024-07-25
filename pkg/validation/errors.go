package validation

import "errors"

var (
	ErrInvalidEmail             = errors.New("provided email is invalid")
	ErrInvalidUsername          = errors.New("provided username is invalid")
	ErrIsRequired               = errors.New("value is required")
	ErrValueTooShort            = errors.New("value is too short")
	ErrValueTooLong             = errors.New("value is too long")
	ErrPasswordDontMatchConfirm = errors.New("the two passwords must match")
)
