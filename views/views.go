package views

import (
	"embed"
	"html/template"
	"io"
	"net/http"

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
		Layout:    BaseLayout,
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

type InputData struct {
	Invalid  bool
	InvalidMsg      string
	OldValue any
}

type RegisterUserData struct {
	NameInput       InputData
	EmailInput      InputData
	PasswordInput   InputData
	ConfirmPassword InputData
}

func (v Views) RegisterUser(ctx echo.Context, data RegisterUserData) error {
	return ctx.Render(http.StatusOK, "user/register", RenderOpts{
		Data: data,
	})
}

func (v Views) RegisteredUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/__registered", RenderOpts{
		Data: nil,
	})
}
