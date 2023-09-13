package services

import (
	"github.com/MBvisti/grafto/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

var (
	passwordPepper = config.GetPwdPepper()
)

func hashAndPepperPassword(password string) (string, error) {
	passwordBytes := []byte(password + passwordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}
