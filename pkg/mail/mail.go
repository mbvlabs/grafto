package mail

import (
	"bytes"
	"context"

	"github.com/mbv-labs/grafto/pkg/mail/templates"
)


type mailClient interface {
	SendMail(ctx context.Context, payload MailPayload) error
}

type Mail struct {
	client mailClient
}

func NewMail(client mailClient) Mail {
	return Mail{
		client: client,
	}
}

func (m *Mail) Send(ctx context.Context, to, from, subject string, tmpl templates.MailTemplateHandler) error {
	var html bytes.Buffer
	if err := tmpl.Render(context.Background(), &html); err != nil {
		return err
	}

	text, err := tmpl.GenerateTextVersion()
	if err != nil {
		return err
	}


	return m.client.SendMail(ctx, MailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: html.String(),
		TextBody: text,
	})
}
