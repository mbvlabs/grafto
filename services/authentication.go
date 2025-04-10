package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/emails"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/routes/paths"
)

func rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		slog.ErrorContext(ctx, "could not rollback transaction", "err", err)
	}
}

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

type emailSender interface {
	Send(
		ctx context.Context,
		payload clients.EmailPayload,
	) error
}

func SendResetPasswordEmail(
	ctx context.Context,
	db psql.Postgres,
	emailClient emailSender,
	email string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)

	user, err := models.GetUserByEmail(
		ctx,
		email,
		tx,
	)
	if err != nil {
		return err
	}

	tkn, err := models.NewToken(
		ctx,
		tx,
		models.NewTokenPayload{
			Expiration: models.ResetPasswordExpirary,
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
			"%s/%s?token=%s",
			config.Cfg.GetFullDomain(),
			paths.GP(ctx, paths.CreateResetPassword),
			tkn.Hash,
		),
	}.Generate(ctx)
	if err != nil {
		return err
	}

	if err := emailClient.Send(ctx, clients.EmailPayload{
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
	hash string,
	password string,
	confirmPassword string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			slog.ErrorContext(ctx, "could not rollback transaction", "err", err)
		}
	}()

	token, err := models.GetToken(
		ctx,
		tx,
		hash,
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
