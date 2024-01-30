package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"time"
)

const (
	ScopeEmailVerification = "email_verification"
	ScopeResetPassword     = "password_reset"
)

type Manager struct {
	hasher hash.Hash
}

func NewManager(tokenSigningKey string) *Manager {
	h := hmac.New(sha256.New, []byte(tokenSigningKey))

	return &Manager{
		h,
	}
}

func (m *Manager) Hash(token string) (string, error) {
	m.hasher.Reset()
	m.hasher.Write([]byte(token))
	b := m.hasher.Sum(nil)

	return base64.URLEncoding.EncodeToString(b), nil
}

func (m *Manager) GenerateToken() (plainText string, hashedToken string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return "", "", err
	}
	plainText = base64.URLEncoding.EncodeToString(b)

	hashedToken, err = m.Hash(plainText)
	if err != nil {
		return "", "", err
	}

	return plainText, hashedToken, nil
}

type Token struct {
	scope     string
	expiresAt time.Time
	Hash      string
	plainText string
}

func (t *Token) GetPlainText() string {
	return t.plainText
}

func (t *Token) GetExpirationTime() time.Time {
	return t.expiresAt
}

func (t *Token) GetScope() string {
	return t.scope
}

func CreateActivationToken(token, hashedToken string) Token {
	return Token{
		ScopeEmailVerification,
		time.Now().Add(72 * time.Hour),
		hashedToken,
		token,
	}
}

func CreateResetPasswordToken(token, hashedToken string) Token {
	return Token{
		ScopeResetPassword,
		time.Now().Add(2 * time.Hour),
		hashedToken,
		token,
	}
}
