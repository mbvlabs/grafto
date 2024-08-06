package emails

import (
	"context"
	"embed"
	"io"
)

//go:embed *.txt
var TextTemplates embed.FS

//go:embed *_templ.go
var HtmlTemplates embed.FS

type TemplateHandler interface {
	GenerateTextVersion() (string, error)
	GenerateHtmlVersion() (string, error)
	Render(ctx context.Context, w io.Writer) error
}
