package services

import (
	"context"
	"errors"
	"time"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/emails"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
)

type EmailSender interface {
	SendTransaction(
		ctx context.Context,
		payload clients.EmailPayload,
	) error
}

func RegisterUser(
	ctx context.Context,
	db psql.Postgres,
	emailClient EmailSender,
	email string,
	password string,
	confirmPassword string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	//nolint:errcheck //how the setup should be
	defer tx.Rollback(ctx)

	if _, err := models.GetUserByEmail(ctx, tx, email); err == nil {
		return errors.New("user already registred")
	}

	user, err := models.NewUser(ctx, tx, models.NewUserPayload{
		Email: email,
		Password: models.PasswordPair{
			Password:        password,
			ConfirmPassword: confirmPassword,
		},
	})
	if err != nil {
		return err
	}

	codeToken, err := models.NewCodeToken(ctx, tx, models.NewTokenPayload{
		Expiration: time.Now().Add(48 * time.Minute),
		Meta: models.MetaInformation{
			Resource:   models.ResourceUser,
			ResourceID: user.ID,
			Scope:      models.ScopeEmailVerification,
		},
	})
	if err != nil {
		return err
	}

	html, txt, err := emails.SignupWelcome{
		VerificationCode: codeToken.Value,
	}.Generate(ctx)
	if err != nil {
		return err
	}

	if err := emailClient.SendTransaction(ctx, clients.EmailPayload{
		To:       user.Email,
		Subject:  "Welcome onboard!",
		HtmlBody: html.String(),
		TextBody: txt.String(),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func ValidateUserEmail(
	ctx context.Context,
	db psql.Postgres,
	tokenValue string,
) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	//nolint:errcheck //how the setup should be
	defer tx.Rollback(ctx)

	token, err := models.GetHashedToken(ctx, tx, tokenValue)
	if err != nil {
		return err
	}

	if !token.IsValid() || token.Meta.Scope != models.ScopeEmailVerification {
		return errors.New("invalid token")
	}

	user, err := models.GetUser(
		ctx,
		tx,
		token.Meta.ResourceID,
	)
	if err != nil {
		return err
	}

	if err := models.UpdateUserEmailToVerified(
		ctx,
		tx,
		models.UpdateUserEmailToVerifiedPayload{
			ID:         user.ID,
			Email:      user.Email,
			VerifiedAt: time.Now(),
		},
	); err != nil {
		return err
	}

	if err := models.DeleteToken(ctx, tx, token.ID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
