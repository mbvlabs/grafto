package main

import (
	"crypto/rand"
	"encoding/base32"
	"log/slog"
	"math/big"
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

	// signupWelcome := emails.SignupWelcome{
	// 	VerificationCode: "d4sss13",
	// }
	// signupWelcomeHtml, signupWelcomeText, _ := signupWelcome.Generate(
	// 	context.Background(),
	// )
	// if err := emailClient.SendTransaction(context.Background(), clients.EmailPayload{
	// 	To:       "vanilla.beetle.spider@aboutmy.email",
	// 	From:     "Morten <noreply@golangblogcourse.com>",
	// 	Subject:  "Tester",
	// 	HtmlBody: signupWelcomeHtml.String(),
	// 	TextBody: signupWelcomeText.String(),
	// }); err != nil {
	// 	slog.Error("EEEEEEEEEEEEEEEEERROR", "e", err)
	// }
}
