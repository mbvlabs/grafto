package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ConfirmEmailClaim struct {
	ConfirmationID uuid.UUID `json:"confirmation_id"`
	jwt.RegisteredClaims
}

func (c *ConfirmEmailClaim) GetSignedJWT(tokenSigningKey string) (string, error) {
	if c.ConfirmationID.String() == "" {
		return "", errors.New("empty uuid provided ConfirmEmailClaim")
	}

	exirationDate := jwt.NewNumericDate(time.Now().Add(48 * time.Hour))

	claims := ConfirmEmailClaim{
		ConfirmationID: c.ConfirmationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: exirationDate,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString(tokenSigningKey)

	return signedToken, nil
}

func (r *ConfirmEmailClaim) ParseJWT(token, tokenSigningKey string) (*ConfirmEmailClaim, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&ConfirmEmailClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return tokenSigningKey, nil
		},
	)

	if claims, ok := parsedToken.Claims.(*ConfirmEmailClaim); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, err
}

type ResetPasswordClaim struct {
	ResetID uuid.UUID `json:"reset_id"`
	jwt.RegisteredClaims
}

func (r *ResetPasswordClaim) Create(id uuid.UUID, tokenSigningKey string) string {
	exirationDate := jwt.NewNumericDate(time.Now().Add(1 * time.Hour))

	claims := ResetPasswordClaim{
		ResetID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: exirationDate,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString(tokenSigningKey)

	return signedToken
}

func (r *ResetPasswordClaim) Parse(token, tokenSigningKey string) (*ResetPasswordClaim, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&ResetPasswordClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return tokenSigningKey, nil
		},
	)

	if claims, ok := parsedToken.Claims.(*ResetPasswordClaim); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, err
}
