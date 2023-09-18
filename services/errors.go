package services

import "github.com/pkg/errors"

var (
	ErrInvalidInput = errors.New("one or more of the provided inputs are not valid")
)
