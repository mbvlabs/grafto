package main

import (
	"context"
	"os"

	"github.com/mbv-labs/grafto/pkg/mail"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
)

func main() {
	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	tmpl := templates.ConfirmUserEmail{
		ConfirmLink:     "https://mortenvistisen.com",
		UnsubscribeLink: "https://mortenvistisen.com",
	}

	if err := mailClient.Send(
		context.Background(), 
		"mbv@mortenvistisen.com", 
		"mbv@mortenvistisen.com", 
		"test new mail", 
		tmpl,
	); err != nil {
		panic(err)
	}
}
