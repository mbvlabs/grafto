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

var DashboardHome = Route{
	Name:         dashboardNamePrefix + ".home",
	Path:         dashboardRoutePrefix,
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "Index",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}

var DashboardUsers = Route{
	Name:         dashboardNamePrefix + ".users",
	Path:         dashboardRoutePrefix + "/users",
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "UsersList",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}

var DashboardUserEdit = Route{
	Name:         dashboardNamePrefix + ".user.edit",
	Path:         dashboardRoutePrefix + "/users/:id/edit",
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "EditUser",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}

var DashboardUserUpdate = Route{
	Name:         dashboardNamePrefix + ".user.update",
	Path:         dashboardRoutePrefix + "/users/:id",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "UpdateUser",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}

var DashboardUserDelete = Route{
	Name:         dashboardNamePrefix + ".user.delete",
	Path:         dashboardRoutePrefix + "/users/:id/delete",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "DeleteUser",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}

var DashboardUserToggleAdmin = Route{
	Name:         dashboardNamePrefix + ".user.toggle_admin",
	Path:         dashboardRoutePrefix + "/users/:id/toggle-admin",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "MakeUserAdmin",
	Middleware: []func(next echo.HandlerFunc) echo.HandlerFunc{
		middleware.AuthOnly,
	},
}
