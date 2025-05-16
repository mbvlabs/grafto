package routes

import (
	"net/http"
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
	Middleware: []string{
		"AuthOnly",
	},
}
