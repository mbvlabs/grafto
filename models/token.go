package models

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/models/internal/db"
)

type (
	Scope    string
	Resource string
)

var (
	ScopeEmailVerification Scope = "email_verification"
	ScopeUnsubscribe       Scope = "unsubscribe"

	ResetPasswordExpirary time.Time = time.Now().Add(1 * time.Hour)
	ScopeResetPassword    Scope     = "password_reset"

	ResourceUser       Resource = "users"
	ResourceSubscriber Resource = "subscribers"
)

type MetaInformation struct {
	Resource   Resource  `validate:"required"`
	ResourceID uuid.UUID `validate:"required,uuid"`
	Scope      Scope     `validate:"required"`
}

type Token struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	Expiration time.Time
	Hash       string
	Meta       MetaInformation
}

func (te Token) IsValid() bool {
	return time.Now().Before(te.Expiration)
}

type NewTokenPayload struct {
	Expiration time.Time       `validate:"required"`
	Meta       MetaInformation `validate:"required"`
}

func NewToken(
	ctx context.Context,
	dbtx db.DBTX,
	data NewTokenPayload,
) (Token, error) {
	if err := validate.Struct(data); err != nil {
		return Token{}, errors.Join(ErrDomainValidation, err)
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return Token{}, err
	}

	hash := base64.URLEncoding.EncodeToString(b)

	now := time.Now()
	tkn := Token{
		ID:         uuid.New(),
		CreatedAt:  now,
		Expiration: data.Expiration,
		Hash:       hash,
		Meta:       data.Meta,
	}

	metaData, err := json.Marshal(data.Meta)
	if err != nil {
		return Token{}, err
	}

	_, err = db.Stmts.InsertToken(ctx, dbtx, db.InsertTokenParams{
		ID: tkn.ID,
		CreatedAt: pgtype.Timestamptz{
			Time:  tkn.CreatedAt,
			Valid: true,
		},
		Hash: hash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  tkn.Expiration,
			Valid: true,
		},
		MetaInformation: metaData,
	})
	if err != nil {
		return Token{}, err
	}

	return tkn, nil
}

func GetToken(
	ctx context.Context,
	dbtx db.DBTX,
	token string,
) (Token, error) {
	tkn, err := db.Stmts.QueryTokenByHash(ctx, dbtx, token)
	if err != nil {
		return Token{}, err
	}

	var meta MetaInformation
	if err := json.Unmarshal(tkn.MetaInformation, &meta); err != nil {
		return Token{}, err
	}

	return Token{
		ID:         tkn.ID,
		CreatedAt:  tkn.CreatedAt.Time,
		Expiration: tkn.ExpiresAt.Time,
		Hash:       tkn.Hash,
		Meta:       meta,
	}, nil
}

func DeleteToken(
	ctx context.Context,
	dbtx db.DBTX,
	tokenID uuid.UUID,
) error {
	err := db.Stmts.DeleteToken(ctx, dbtx, tokenID)
	if err != nil {
		return err
	}

	return nil
}
