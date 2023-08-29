package mail

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"
)

type mailClient interface {
	SendMail(ctx context.Context, payload MailPayload) error
}

//go:embed templates/*.html
var templates embed.FS

type Mail struct {
	client mailClient
}

func NewMail(client mailClient) Mail {
	return Mail{
		client: client,
	}
}

func (m *Mail) Send(to, from, subject, tmpl string, data interface{}) error {
	t, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.html", tmpl))
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return err
	}

	return m.client.SendMail(context.Background(), MailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: tpl.String(),
		TextBody: "ignore",
	})
}
