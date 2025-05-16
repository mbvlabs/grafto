package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
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

	ScopeResetPassword Scope = "password_reset"

	ResourceUser       Resource = "users"
	ResourceSubscriber Resource = "subscribers"
)

type MetaInformation struct {
	Resource   Resource  `validate:"required"      json:"resource"`
	ResourceID uuid.UUID `validate:"required,uuid" json:"resource_id"`
	Scope      Scope     `validate:"required"      json:"scope"`
}

type Token struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	Expiration time.Time
	Value      string
	Meta       MetaInformation
}

func (te Token) IsValid() bool {
	return time.Now().Before(te.Expiration)
}

type NewTokenPayload struct {
	Expiration time.Time       `validate:"required"`
	Meta       MetaInformation `validate:"required" json:"meta"`
}

func generateToken() string {
	bytes := make([]byte, 15)
	//nolint:errcheck //can't error
	rand.Read(bytes)
	return strings.ToLower(base32.StdEncoding.EncodeToString(bytes))
}

func generateHash(token string) string {
	hash := sha256.New()

	hash.Write([]byte(strings.ToLower(token)))

	hashedToken := hash.Sum(nil)

	return hex.EncodeToString(hashedToken)
}

func generateRandomAlphanumeric(length int) (string, error) {
	const charset = "abcdefghjklmnpqrstuvwxyz23456789"
	result := make([]byte, length)
	for i := range result {
		randomIndex, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charset))),
		)
		if err != nil {
			return "", err
		}
		result[i] = charset[randomIndex.Int64()]
	}
	return string(result), nil
}

func NewToken(
	ctx context.Context,
	dbtx db.DBTX,
	data NewTokenPayload,
) (Token, error) {
	if err := validate.Struct(data); err != nil {
		return Token{}, errors.Join(ErrDomainValidation, err)
	}

	return newToken(ctx, dbtx, data.Expiration, data.Meta, generateToken())
}

func NewHashedToken(
	ctx context.Context,
	dbtx db.DBTX,
	data NewTokenPayload,
) (Token, error) {
	if err := validate.Struct(data); err != nil {
		return Token{}, errors.Join(ErrDomainValidation, err)
	}

	tkn := generateToken()
	hashedToken := generateHash(tkn)

	newToken, err := newToken(
		ctx,
		dbtx,
		data.Expiration,
		data.Meta,
		hashedToken,
	)
	if err != nil {
		return Token{}, err
	}

	return Token{
		ID:         newToken.ID,
		CreatedAt:  newToken.CreatedAt,
		Expiration: newToken.Expiration,
		Value:      tkn,
		Meta:       newToken.Meta,
	}, nil
}

func NewCodeToken(
	ctx context.Context,
	dbtx db.DBTX,
	data NewTokenPayload,
) (Token, error) {
	if err := validate.Struct(data); err != nil {
		return Token{}, errors.Join(ErrDomainValidation, err)
	}

	codeTkn, err := generateRandomAlphanumeric(6)
	if err != nil {
		return Token{}, err
	}

	newToken, err := newToken(
		ctx,
		dbtx,
		data.Expiration,
		data.Meta,
		generateHash(codeTkn),
	)
	if err != nil {
		return Token{}, err
	}

	return Token{
		ID:         newToken.ID,
		CreatedAt:  newToken.CreatedAt,
		Expiration: newToken.Expiration,
		Value:      codeTkn,
		Meta:       newToken.Meta,
	}, nil
}

func newToken(
	ctx context.Context,
	dbtx db.DBTX,
	expiration time.Time,
	meta MetaInformation,
	token string,
) (Token, error) {
	now := time.Now()

	tkn := Token{
		ID:         uuid.New(),
		CreatedAt:  now,
		Expiration: expiration,
		Value:      token,
		Meta:       meta,
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		return Token{}, err
	}

	_, err = db.Stmts.InsertToken(ctx, dbtx, db.InsertTokenParams{
		ID: tkn.ID,
		CreatedAt: pgtype.Timestamptz{
			Time:  tkn.CreatedAt,
			Valid: true,
		},
		Hash: tkn.Value,
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
	tokenRow, err := db.Stmts.QueryTokenByHash(ctx, dbtx, token)
	if err != nil {
		return Token{}, err
	}

	var meta MetaInformation
	if err := json.Unmarshal(tokenRow.MetaInformation, &meta); err != nil {
		return Token{}, err
	}

	return Token{
		ID:         tokenRow.ID,
		CreatedAt:  tokenRow.CreatedAt.Time,
		Expiration: tokenRow.ExpiresAt.Time,
		Value:      tokenRow.Hash,
		Meta:       meta,
	}, nil
}

func GetHashedToken(
	ctx context.Context,
	dbtx db.DBTX,
	token string,
) (Token, error) {
	tokenRow, err := db.Stmts.QueryTokenByHash(ctx, dbtx, generateHash(token))
	if err != nil {
		return Token{}, err
	}

	var meta MetaInformation
	if err := json.Unmarshal(tokenRow.MetaInformation, &meta); err != nil {
		return Token{}, err
	}

	return Token{
		ID:         tokenRow.ID,
		CreatedAt:  tokenRow.CreatedAt.Time,
		Expiration: tokenRow.ExpiresAt.Time,
		Value:      tokenRow.Hash,
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
