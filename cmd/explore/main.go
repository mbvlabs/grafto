package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"log/slog"
	"math/big"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/emails"
)

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

func generateToken() string {
	bytes := make([]byte, 15)
	//nolint:errcheck //can't error
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func main() {
	slog.Info("explore")
	// Create a new SHA-256 hash
	// hash := sha256.New()

	num, _ := generateRandomAlphanumeric(6)
	// Write the token data to the hash
	// hash.Write([]byte(num))
	//
	// // Get the resulting hash as a byte slice
	// hashedToken := hash.Sum(nil)
	//
	// // Convert the byte slice to a hexadecimal string
	// hashedTokenString := hex.EncodeToString(hashedToken)
	//
	h, t, e := emails.SignupWelcome{
		VerificationCode: num,
	}.Generate(context.Background())
	if e != nil {
		panic(e)
	}

	ec := clients.NewEmail()
	if err := ec.Send(context.Background(), clients.EmailPayload{
		To:       "hi@mbvlabs.com",
		From:     "info@golangblogcourse.com",
		Subject:  "Welcome onboard",
		HtmlBody: h.String(),
		TextBody: t.String(),
	}, clients.Unsubscribe{}); err != nil {
		panic(err)
	}
}
