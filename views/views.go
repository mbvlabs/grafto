package views

import (
	"embed"
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"github.com/unrolled/render"
)

//go:embed templates/***/*.html
var templates embed.FS

var (
	BaseLayout      = "layouts/base"
	DashboardLayout = "layouts/dashboard"
)

type Views struct {
	render *render.Render
}

func NewViews() Views {
	r := render.New(render.Options{
		Directory: "templates",
		// Layout:    BaseLayout,
		FileSystem: &render.EmbedFileSystem{
			FS: templates,
		},
		Extensions:      []string{".html"},
		RequirePartials: false,
		Funcs: []template.FuncMap{
			sprig.FuncMap(),
		},
	})

	return Views{
		render: r,
	}
}

type RenderOpts struct {
	Layout string
	Data   interface{}
}

func (v Views) Render(w io.Writer, tmpl string, data interface{}, e echo.Context) error {
	renderOpts, ok := data.(RenderOpts)
	if data != nil && !ok {
		panic("bad render opts") // TODO add fallback tmpl
	}

	if renderOpts.Layout != "" {
		return v.render.HTML(w, 0, tmpl, renderOpts.Data, render.HTMLOptions{
			Layout: renderOpts.Layout,
		})
	}
	return v.render.HTML(w, 0, tmpl, renderOpts.Data)
}
