package views

import (
	"embed"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/MBvisti/grafto/routes/middleware"
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
		FileSystem: &render.EmbedFileSystem{
			FS: templates,
		},
		Extensions:      []string{".html"},
		RequirePartials: false,
		Funcs: []template.FuncMap{
			sprig.FuncMap(),
		},
		IsDevelopment: os.Getenv("ENVIRONMENT") == "development",
	})

	return Views{
		render: r,
	}
}

type auth struct {
	IsAuthenticated bool
	IsAdmin         bool
}

type RenderOpts struct {
	Layout string
	Data   interface{}
	Auth   auth
}

func (v Views) Render(w io.Writer, tmpl string, data interface{}, e echo.Context) error {
	renderOpts, ok := data.(RenderOpts)
	if data != nil && !ok {
		panic("bad render opts") // TODO add fallback tmpl
	}

	authContext, ok := e.(*middleware.AuthContext)
	if !ok {
		renderOpts.Auth.IsAuthenticated = false
	} else {
		renderOpts.Auth.IsAuthenticated = authContext.GetAuthStatus()
	}

	adminContext, ok := e.(*middleware.AdminContext)
	if !ok {
		renderOpts.Auth.IsAdmin = false
	} else {
		renderOpts.Auth.IsAdmin = adminContext.GetAdminStatus()
	}

	if renderOpts.Layout != "" {
		return v.render.HTML(w, 0, tmpl, renderOpts, render.HTMLOptions{
			Layout: renderOpts.Layout,
		})
	}

	return v.render.HTML(w, 0, tmpl, renderOpts)
}

type InternalServerErrData struct {
	FromLocation string
}

func (v Views) InternalServerErr(ctx echo.Context, data InternalServerErrData) error {
	return ctx.Render(http.StatusOK, "errors/500", RenderOpts{
		Layout: BaseLayout,
		Data:   data,
	})
}
