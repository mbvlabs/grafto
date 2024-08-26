package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/psql/database"
)

const (
	ScopeEmailVerification = "email_verification"
	ScopeUnsubscribe       = "unsubscribe"
	ScopeResetPassword     = "password_reset"
)

const (
	resourceUser       = "users"
	resourceSubscriber = "subscribers"
)

type TokenMetaInformation struct {
	Resource   string    `json:"resource"`
	ResourceID uuid.UUID `json:"resource_id"`
	Scope      string    `json:"scope"`
}

type tokenServiceStorage interface {
	InsertToken(
		ctx context.Context,
		hash string,
		expiresAt time.Time,
		metaData []byte,
	) error
	QueryTokenByHash(ctx context.Context, hash string) (database.Token, error)
	DeleteTokenByHash(ctx context.Context, hash string) error
}

type Token struct {
	storage tokenServiceStorage
	hasher  hash.Hash
}

func NewTokenSvc(
	storage tokenServiceStorage,
	tokenSigningKey string,
) *Token {
	h := hmac.New(sha256.New, []byte(tokenSigningKey))

	return &Token{
		storage,
		h,
	}
}

type tokenPair struct {
	plain  string
	hashed string
}

func (svc *Token) create() (tokenPair, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return tokenPair{}, err
	}

	plainText := base64.URLEncoding.EncodeToString(b)
	hashedToken := svc.hash(plainText)

	return tokenPair{
		plainText,
		hashedToken,
	}, nil
}

func (svc *Token) hash(token string) string {
	svc.hasher.Reset()
	svc.hasher.Write([]byte(token))
	b := svc.hasher.Sum(nil)

	return base64.URLEncoding.EncodeToString(b)
}

func (svc *Token) CreateSubscriberEmailValidation(
	ctx context.Context,
	subscriberID uuid.UUID,
) (string, error) {
	tokenPair, err := svc.create()
	if err != nil {
		return "", err
	}

	metaData, err := json.Marshal(TokenMetaInformation{
		Resource:   resourceSubscriber,
		ResourceID: subscriberID,
		Scope:      ScopeEmailVerification,
	})
	if err != nil {
		return "", err
	}

	expirationDate := time.Now().Add(72 * time.Hour)

	if err := svc.storage.InsertToken(ctx, tokenPair.hashed, expirationDate, metaData); err != nil {
		slog.ErrorContext(
			ctx,
			"could not insert a subscriber token",
			"error",
			err,
			"subscriber_id",
			subscriberID,
		)
		return "", err
	}

	return tokenPair.plain, nil
}

func (svc *Token) CreateUserEmailVerification(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	tokenPair, err := svc.create()
	if err != nil {
		return "", err
	}

	metaData, err := json.Marshal(TokenMetaInformation{
		Resource:   resourceUser,
		ResourceID: userID,
		Scope:      ScopeEmailVerification,
	})
	if err != nil {
		return "", err
	}

	expirationDate := time.Now().Add(48 * time.Hour)

	if err := svc.storage.InsertToken(ctx, tokenPair.hashed, expirationDate, metaData); err != nil {
		slog.ErrorContext(
			ctx,
			"could not insert a user email verification token",
			"error",
			err,
			"user_id",
			userID,
		)
		return "", err
	}

	return tokenPair.plain, nil
}

func (svc *Token) CreateResetPasswordToken(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	tokenPair, err := svc.create()
	if err != nil {
		return "", err
	}

	metaData, err := json.Marshal(TokenMetaInformation{
		ResourceID: userID,
		Scope:      ScopeResetPassword,
	})
	if err != nil {
		return "", err
	}

	expirationDate := time.Now().Add(24 * time.Hour)

	if err := svc.storage.InsertToken(ctx, tokenPair.hashed, expirationDate, metaData); err != nil {
		slog.ErrorContext(
			ctx,
			"could not insert a reset password token",
			"error",
			err,
			"user_id",
			userID,
		)
		return "", err
	}

	return tokenPair.plain, nil
}

func (svc *Token) CreateUnsubscribeToken(
	ctx context.Context,
	subscriberID uuid.UUID,
) (string, error) {
	tokenPair, err := svc.create()
	if err != nil {
		return "", err
	}

	metaData, err := json.Marshal(TokenMetaInformation{
		Resource:   resourceSubscriber,
		ResourceID: subscriberID,
		Scope:      ScopeUnsubscribe,
	})
	if err != nil {
		return "", err
	}

	expirationDate := time.Now().Add(168 * time.Hour)

	if err := svc.storage.InsertToken(ctx, tokenPair.hashed, expirationDate, metaData); err != nil {
		slog.ErrorContext(
			ctx,
			"could not insert a unsubscribe token",
			"error",
			err,
			"subscriber_id",
			subscriberID,
		)
		return "", err
	}

	return tokenPair.plain, nil
}

func (svc *Token) Validate(ctx context.Context, token, scope string) error {
	tkn, err := svc.storage.QueryTokenByHash(ctx, svc.hash(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(ctx, "a token was requested that could not be found", "token", tkn)

			slog.ErrorContext(ctx, "could not query token by hash", "error", err)
			return errors.Join(ErrTokenNotExist, err)
		}

		return err
	}

	if time.Now().After(tkn.ExpiresAt.Time) {
		return ErrTokenExpired
	}

	var metaInfo TokenMetaInformation
	if err := json.Unmarshal(tkn.MetaInformation, &metaInfo); err != nil {
		return err
	}

	if metaInfo.Scope != scope {
		return ErrTokenScopeInvalid
	}

	return nil
}

func (svc *Token) IsExpired(ctx context.Context, token string) error {
	tkn, err := svc.storage.QueryTokenByHash(ctx, svc.hash(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(ctx, "a token was requested that could not be found", "token", tkn)

			slog.ErrorContext(ctx, "could not query token by hash", "error", err)
			return errors.Join(ErrTokenNotExist, err)
		}

		return err
	}

	if time.Now().After(tkn.ExpiresAt.Time) {
		return ErrTokenExpired
	}

	return nil
}

func (svc *Token) GetAssociatedUserID(ctx context.Context, token string) (uuid.UUID, error) {
	tkn, err := svc.storage.QueryTokenByHash(ctx, svc.hash(token))
	if err != nil {
		return uuid.UUID{}, err
	}

	var metaData TokenMetaInformation
	if err := json.Unmarshal(tkn.MetaInformation, &metaData); err != nil {
		return uuid.UUID{}, err
	}

	// TODO add err here
	if metaData.Resource != resourceUser {
		return uuid.UUID{}, err
	}

	return metaData.ResourceID, nil
}

func (svc *Token) GetAssociatedSubscriberID(
	ctx context.Context,
	token string,
) (uuid.UUID, error) {
	tkn, err := svc.storage.QueryTokenByHash(ctx, svc.hash(token))
	if err != nil {
		return uuid.UUID{}, err
	}

	var metaData TokenMetaInformation
	if err := json.Unmarshal(tkn.MetaInformation, &metaData); err != nil {
		return uuid.UUID{}, err
	}

	// TODO add err here
	if metaData.Resource != resourceSubscriber {
		return uuid.UUID{}, err
	}

	return metaData.ResourceID, nil
}

func (svc *Token) Delete(ctx context.Context, token string) error {
	err := svc.storage.DeleteTokenByHash(ctx, svc.hash(token))
	if err != nil {
		return err
	}

	return nil
}
