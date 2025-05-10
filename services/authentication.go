package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/emails"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/routes"
)

var (
	ErrUserEmailNotVerified = errors.New("user email is not verified")
	ErrInvalidAuthDetail    = errors.New(
		"the provided details does not match our records",
	)
	ErrInvalidResetToken = errors.New("provided token is invalid")
)

func AuthenticateUser(
	ctx context.Context,
	db psql.Postgres,
	email string,
	providedPassword string,
) (models.UserEntity, error) {
	user, err := models.GetUserByEmail(
		ctx,
		db.Pool,
		email,
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

func SendResetPasswordEmail(
	ctx context.Context,
	db psql.Postgres,
	emailClient EmailSender,
	email string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	//nolint:errcheck //how the setup should be
	defer tx.Rollback(ctx)

	user, err := models.GetUserByEmail(
		ctx,
		tx,
		email,
	)
	if err != nil {
		return err
	}

	tkn, err := models.NewHashedToken(
		ctx,
		tx,
		models.NewTokenPayload{
			Expiration: time.Now().Add(1 * time.Hour),
			Meta: models.MetaInformation{
				Resource:   models.ResourceUser,
				ResourceID: user.ID,
				Scope:      models.ScopeResetPassword,
			},
		},
	)
	if err != nil {
		return err
	}

	html, txt, err := emails.PasswordReset{
		ResetLink: fmt.Sprintf(
			"%s%s?token=%s",
			config.Cfg.GetFullDomain(),
			routes.ResetPasswordPage.Path,
			tkn.Value,
		),
	}.Generate(ctx)
	if err != nil {
		return err
	}

	if err := emailClient.SendTransaction(ctx, clients.EmailPayload{
		To:       user.Email,
		Subject:  "Action Required | Password reset requested",
		HtmlBody: html.String(),
		TextBody: txt.String(),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func ChangeUserPassword(
	ctx context.Context,
	db psql.Postgres,
	providedToken string,
	password string,
	confirmPassword string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	//nolint:errcheck //how the setup should be
	defer tx.Rollback(ctx)

	token, err := models.GetHashedToken(
		ctx,
		tx,
		providedToken,
	)
	if err != nil {
		return err
	}

	if !token.IsValid() || token.Meta.Scope != models.ScopeResetPassword {
		return ErrInvalidResetToken
	}

	if err := models.UpdateUserPassword(
		ctx,
		tx,
		models.UpdateUserPasswordPayload{
			ID:        token.Meta.ResourceID,
			UpdatedAt: time.Now(),
			Password: models.PasswordPair{
				Password:        password,
				ConfirmPassword: confirmPassword,
			},
		},
	); err != nil {
		return err
	}

	if err := models.DeleteToken(
		ctx, tx, token.ID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
