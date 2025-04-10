package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/emails"
)

func main() {
	slog.Info("explore")
	emailClient := clients.NewEmail()

	signupWelcome := emails.SignupWelcome{
		ConfirmationLink: fmt.Sprintf(
			"%s/%s?token=%s",
			config.Cfg.GetFullDomain(),
			"reset-password",
			"wvSwI8Yq02o9cmJ6zVSTkP44lXGJZjmMF8v10vxAhrrV6UyzRr59ogUzdo3VKP7y",
		),
	}
	signupWelcomeHtml, signupWelcomeText, _ := signupWelcome.Generate(
		context.Background(),
	)
	if err := emailClient.Send(context.Background(), clients.EmailPayload{
		To:       "heymbv@gmail.com",
		From:     "noreply@golangblogcourse.com",
		Subject:  "Tester",
		HtmlBody: signupWelcomeHtml.String(),
		TextBody: signupWelcomeText.String(),
	}); err != nil {
		slog.Error("EEEEEEEEEEEEEEEEERROR", "e", err)
	}
	if err := emailClient.Send(context.Background(), clients.EmailPayload{
		To:       "morten@mbvlabs.com",
		From:     "noreply@golangblogcourse.com",
		Subject:  "Tester",
		HtmlBody: signupWelcomeHtml.String(),
		TextBody: signupWelcomeText.String(),
	}); err != nil {
		slog.Error("EEEEEEEEEEEEEEEEERROR", "e", err)
	}
}
