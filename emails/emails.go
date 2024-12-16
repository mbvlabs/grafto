package emails

import (
	"bytes"
	"context"
	"embed"

	"github.com/a-h/templ"
	"github.com/jaytaylor/html2text"
	"github.com/vanng822/go-premailer/premailer"
)

type (
	Html string
	Text string
)

func (h Html) String() string {
	return string(h)
}

func (t Text) String() string {
	return string(t)
}

//go:embed *_templ.go
var HtmlTemplates embed.FS

type TemplateHandler interface {
	Generate(ctx context.Context) (Html, Text, error)
}

func processEmail(ctx context.Context, tmpl templ.Component) (string, string, error) {
	var html bytes.Buffer
	if err := tmpl.Render(ctx, &html); err != nil {
		return "", "", err
	}

	premailer, err := premailer.NewPremailerFromString(html.String(), premailer.NewOptions())
	if err != nil {
		return "", "", err
	}

	inlineHtml, err := premailer.Transform()
	if err != nil {
		return "", "", err
	}

	plainText, err := html2text.FromString(inlineHtml, html2text.Options{PrettyTables: false})
	if err != nil {
		return "", "", err
	}

	return inlineHtml, plainText, nil
}
