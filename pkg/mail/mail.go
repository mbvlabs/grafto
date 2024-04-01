package mail

import (
	"context"
)

type MailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

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

func (m *Mail) Send(
	ctx context.Context,
	to,
	from,
	subject,
	textVersion,
	htmlVersion string,
) error {
	return m.client.SendMail(ctx, MailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: htmlVersion,
		TextBody: textVersion,
	})
}
