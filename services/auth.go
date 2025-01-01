package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/emails"
	"github.com/mbvlabs/grafto/models"
	emailClient "github.com/mbvlabs/grafto/pkg/email_client"
	"github.com/mbvlabs/grafto/psql"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	db    psql.Postgres
	email emailClient.EmailClient
}

func NewAuth(
	db psql.Postgres,
	email emailClient.EmailClient,
) Auth {
	return Auth{db, email}
}

// TODO: maybe this should be moved
func HashAndPepperPassword(password string) (string, error) {
	passwordBytes := []byte(password + config.Cfg.PasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(
		passwordBytes,
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func validatePassword(password string) error {
	hashedPassword, err := HashAndPepperPassword(password)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password+config.Cfg.PasswordPepper),
	)
}

func (a Auth) AuthenticateUser(
	ctx context.Context,
	email string,
	password string,
) (models.UserEntity, error) {
	user, err := models.GetUserByEmail(ctx, email, a.db.Pool)
	if err != nil {
		slog.ErrorContext(ctx, "could not query user", "error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserEntity{}, ErrUserNotExist
		}

		return models.UserEntity{}, err
	}

	if isVerified := user.IsVerified(); !isVerified {
		return models.UserEntity{}, ErrEmailNotValidated
	}

	if err := validatePassword(password); err != nil {
		return models.UserEntity{}, ErrPasswordNotMatch
	}

	return user, nil
}

func (a Auth) RegisterUser(
	ctx context.Context,
	name string,
	email string,
	password string,
	confirmPassword string,
) error {
	tx, err := a.db.BeginTx(ctx)
	if err != nil {
		return errors.Join(ErrUnrecoverable, err)
	}

	user, err := models.NewUser(ctx, models.NewUserPayload{
		Email:    email,
		Password: password,
	}, tx, HashAndPepperPassword)
	if err != nil {
		if !errors.Is(err, models.ErrDomainValidation) {
			return errors.Join(ErrUnrecoverable, err)
		}

		return err
	}

	html, text, err := emails.SignupWelcome{
		ConfirmationLink: fmt.Sprintf(
			"%s/verify-email?token=%s",
			config.Cfg.GetFullDomain(),
			"",
		),
	}.Generate(ctx)
	if err != nil {
		return err
	}

	if err := a.email.Send(ctx, user.Email, config.Cfg.DefaultSenderSignature, "Grafto | Action Required", html.String(), text.String()); err != nil {
		return errors.Join(ErrUnrecoverable, err)
	}

	return nil
}

func (a Auth) VerifyUserEmail(
	ctx context.Context,
	token string,
	scope string,
) error {
	tx, err := a.db.BeginTx(ctx)
	if err != nil {
		return errors.Join(ErrUnrecoverable, err)
	}
	defer tx.Rollback(ctx)

	tkn, err := models.GetToken(ctx, token, tx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenNotExist
		}

		return errors.Join(ErrUnrecoverable, err)
	}

	if tkn.Meta.Scope != models.ScopeEmailVerification {
		return ErrTokenScopeInvalid
	}

	if !tkn.IsValid() {
		return ErrTokenExpired
	}

	user, err := models.GetUser(ctx, tkn.Meta.ResourceID, tx)
	if err != nil {
		return errors.Join(ErrUnrecoverable, err)
	}

	if _, err := models.UpdateUser(ctx, models.UpdateUserPayload{
		ID:             user.ID,
		UpdatedAt:      time.Now(),
		Email:          user.Email,
		EmailUpdatedAt: time.Now(),
	}, tx); err != nil {
		return err
	}

	// if _, err := models.DeleteToken(); err != nil {
	// 	return err
	// }

	if err := tx.Commit(ctx); err != nil {
		return errors.Join(ErrUnrecoverable, err)
	}

	return nil
}
