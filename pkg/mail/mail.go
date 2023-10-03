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

//go:embed templates/*
var templates embed.FS

type Mail struct {
	client mailClient
}

func NewMail(client mailClient) Mail {
	return Mail{
		client: client,
	}
}

func (m *Mail) Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error {
	htmlFile, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.html", tmplName))
	if err != nil {
		return err
	}

	var htmlBody bytes.Buffer
	if err := htmlFile.Execute(&htmlBody, data); err != nil {
		return err
	}

	textFile, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.txt", tmplName))
	if err != nil {
		return err
	}

	var textBody bytes.Buffer
	if err := textFile.Execute(&textBody, data); err != nil {
		return err
	}

	return m.client.SendMail(ctx, MailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: htmlBody.String(),
		TextBody: textBody.String(),
	})
}
