package services

import (
	"context"
	"errors"

	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
)

var (
	ErrUserEmailNotVerified = errors.New("user email is not verified")
	ErrInvalidAuthDetail    = errors.New(
		"the provided details does not match our records",
	)
)

func AuthenticateUser(
	ctx context.Context,
	db psql.Postgres,
	email string,
	providedPassword string,
) (models.UserEntity, error) {
	user, err := models.GetUserByEmail(
		ctx,
		email,
		db.Pool,
	)
	if err != nil {
		return models.UserEntity{}, ErrInvalidAuthDetail
	}

	if !user.IsVerified() {
		return models.UserEntity{}, ErrUserEmailNotVerified
	}

	if err := user.ValidatePassword(providedPassword); err != nil {
		return models.UserEntity{}, ErrInvalidAuthDetail
	}

	return user, nil
}
