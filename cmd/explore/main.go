package main

import (
	"log/slog"
)

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
