package models

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/internal/db"
)

var Token TokenEntity

var h = hmac.New(sha256.New, []byte(config.Cfg.TokenSigningKey))

type (
	Scope    string
	Resource string
)

var (
	ScopeEmailVerification Scope = "email_verification"
	ScopeUnsubscribe       Scope = "unsubscribe"
	ScopeResetPassword     Scope = "password_reset"
)

var (
	ResourceUser       Resource = "users"
	ResourceSubscriber Resource = "subscribers"
)

type MetaInformation struct {
	Resource   Resource  `validate:"required"`
	ResourceID uuid.UUID `validate:"required,uuid"`
	Scope      Scope     `validate:"required"`
}

type TokenEntity struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	Expiration time.Time
	Hash       string
	Meta       MetaInformation
}

func (te TokenEntity) IsValid() bool {
	return time.Now().After(te.Expiration)
}

type NewTokenPayload struct {
	Expiration time.Time       `validate:"required"`
	Meta       MetaInformation `validate:"required"`
}

func NewToken(
	ctx context.Context,
	data NewTokenPayload,
	dbtx db.DBTX,
) (TokenEntity, error) {
	if err := validate.Struct(data); err != nil {
		return TokenEntity{}, errors.Join(ErrDomainValidation, err)
	}

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return TokenEntity{}, err
	}

	plainText := base64.URLEncoding.EncodeToString(b)
	h.Reset()
	h.Write([]byte(plainText))
	bytes := h.Sum(nil)

	hash := base64.URLEncoding.EncodeToString(bytes)

	tkn := TokenEntity{
		uuid.New(),
		time.Now(),
		data.Expiration,
		hash,
		data.Meta,
	}

	metaData, err := json.Marshal(data.Meta)
	if err != nil {
		return TokenEntity{}, err
	}

	_, err = db.Stmts.InsertToken(ctx, dbtx, db.InsertTokenParams{
		ID: tkn.ID,
		CreatedAt: pgtype.Timestamptz{
			Time:  tkn.CreatedAt,
			Valid: true,
		},
		Hash: hash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  tkn.CreatedAt,
			Valid: true,
		},
		MetaInformation: metaData,
	})
	if err != nil {
		return TokenEntity{}, err
	}

	return tkn, nil
}

func GetToken(
	ctx context.Context,
	token string,
	dbtx db.DBTX,
) (TokenEntity, error) {
	tkn, err := db.Stmts.QueryTokenByHash(ctx, dbtx, token)
	if err != nil {
		return TokenEntity{}, err
	}

	var meta MetaInformation
	if err := json.Unmarshal(tkn.MetaInformation, &meta); err != nil {
		return TokenEntity{}, err
	}

	return TokenEntity{
		ID:         tkn.ID,
		CreatedAt:  tkn.CreatedAt.Time,
		Expiration: tkn.ExpiresAt.Time,
		Hash:       tkn.Hash,
		Meta:       meta,
	}, nil
}
