package main

import (
	"context"
	"log/slog"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/emails"
)

func main() {
	slog.Info("explore")
	emailClient := clients.NewEmail()

	signupWelcome := emails.SignupWelcome{
		VerificationCode: "d4sss13",
	}
	signupWelcomeHtml, signupWelcomeText, _ := signupWelcome.Generate(
		context.Background(),
	)
	if err := emailClient.SendTransaction(context.Background(), clients.EmailPayload{
		To:       "vanilla.beetle.spider@aboutmy.email",
		From:     "Morten <noreply@golangblogcourse.com>",
		Subject:  "Tester",
		HtmlBody: signupWelcomeHtml.String(),
		TextBody: signupWelcomeText.String(),
	}); err != nil {
		slog.Error("EEEEEEEEEEEEEEEEERROR", "e", err)
	}
}
