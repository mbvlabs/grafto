package emailclient

import (
	"context"
)

type EmailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

type emailer interface {
	send(
		ctx context.Context,
		payload EmailPayload,
	) error
}

type EmailClient struct {
	client emailer
}

func NewEmail(
	client emailer,
	// queueClient QueueClient,
) EmailClient {
	return EmailClient{
		client,
	}
}

func (e *EmailClient) Send(
	ctx context.Context,
	to,
	from,
	subject,
	htmlVersion,
	textVersion string,
) error {
	return e.client.send(ctx, EmailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: htmlVersion,
		TextBody: textVersion,
	})
}
