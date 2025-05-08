package services

import (
	"context"

	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
)

func RegisterUser(
	ctx context.Context,
	db psql.Postgres,
	email string,
	password string,
	confirmPassword string,
) error {
	return nil
}

func ValidateUserEmail(
	ctx context.Context,
	db psql.Postgres,
	tokenValue string,
) error {
	_, err := models.GetHashedToken(ctx, db.Pool, tokenValue)
	if err != nil {
		return err
	}

	return nil
}
