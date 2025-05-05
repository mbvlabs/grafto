package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/router/middleware"
)

const (
	dashboardRoutePrefix = "/dashboard"
	dashboardNamePrefix  = "dashboard"
)

var Dashboard = []Route{
	DashboardHome,
}

var DashboardHome = Route{
	Name:     dashboardNamePrefix + ".home",
	Path:     dashboardRoutePrefix,
	Method:   http.MethodGet,
	CtrlName: "Index",
	Middleware: []echo.MiddlewareFunc{
		middleware.AuthOnly,
	},
}
