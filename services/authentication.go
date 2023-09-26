package services

import (
	"context"
	"errors"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/jackc/pgx/v4"
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

type validatePasswordPayload struct {
	hashedpassword string
	password       string
}

func validatePassword(data validatePasswordPayload) error {
	return bcrypt.CompareHashAndPassword([]byte(data.hashedpassword), []byte(data.password+passwordPepper))
}

type AuthenticateUserPayload struct {
	Email    string
	Password string
}

func AuthenticateUser(ctx context.Context, data AuthenticateUserPayload, db userDatabase) (entity.User, error) {
	user, err := db.QueryUserByMail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, ErrUserNotExist
		}

		telemetry.Logger.ErrorContext(ctx, "could not query user", "error", err)
		return entity.User{}, err
	}

	if verifiedAt := user.MailVerifiedAt; !verifiedAt.Valid {
		return entity.User{}, ErrEmailNotValidated
	}

	err = validatePassword(validatePasswordPayload{
		hashedpassword: user.Password,
		password:       data.Password,
	})
	if err != nil {
		return entity.User{}, ErrPasswordNotMatch
	}

	return entity.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
}
