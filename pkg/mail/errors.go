package mail

import "errors"

var (
	ErrCouldNotSend = errors.New("could not send mail")
	ErrNotAuthorized = errors.New("Unauthorized")
)
