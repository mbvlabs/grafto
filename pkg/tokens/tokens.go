package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
)

var tokenSigningKey = []byte(os.Getenv("TOKEN_SIGNING_KEY"))

const (
	ScopeEmailVerification = "email_verification"
	ScopeResetPassword     = "password_reset"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func HashToken(token string) (string, error) {
	telemetry.Logger.Info("tokens", "token", token)
	h := hmac.New(sha256.New, tokenSigningKey)
	h.Write([]byte(token))
	b := h.Sum(nil)
	telemetry.Logger.Info("tokens", "hashed_token", base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b))
	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

type Token struct {
	scope       string
	expiresAt   time.Time
	HashedToken string
	rawToken    string
}

func (t *Token) GetRawToken() string {
	return t.rawToken
}

func (t *Token) GetExpirationTime() time.Time {
	return t.expiresAt
}

func (t *Token) GetScope() string {
	return t.scope
}

func createToken(scope string, expirationTime time.Time) (Token, error) {
	tkn, err := generateToken()
	if err != nil {
		return Token{}, err
	}

	hashTkn, err := HashToken(tkn)
	if err != nil {
		return Token{}, err
	}

	telemetry.Logger.Info("tokens", "raw", tkn, "hash", hashTkn)

	return Token{
		scope:       scope,
		expiresAt:   expirationTime,
		HashedToken: hashTkn,
		rawToken:    tkn,
	}, err
}

func CreateActivationToken() (Token, error) {
	return createToken(ScopeEmailVerification, time.Now().Add(72*time.Hour))
}

func CreateResetPasswordToken() (Token, error) {
	return createToken(ScopeEmailVerification, time.Now().Add(2*time.Hour))
}
