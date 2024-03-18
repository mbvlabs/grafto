package templates

import (
	"context"
	"embed"
	"io"
)

//go:embed *.txt
var textTemplates embed.FS

type MailTemplateHandler interface {
	GenerateTextVersion() (string, error)
	GenerateHtmlVersion() (string, error)
	Render(ctx context.Context, w io.Writer) error
}
