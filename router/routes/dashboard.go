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
	DashboardUsers,
	DashboardUserEdit,
	DashboardUserUpdate,
	DashboardUserDelete,
	DashboardUserToggleAdmin,
}

var DashboardHome = Route{
	Name:         dashboardNamePrefix + ".home",
	Path:         dashboardRoutePrefix,
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "Index",
	Middleware: []string{
		"AuthOnly",
	},
}

var DashboardUsers = Route{
	Name:         dashboardNamePrefix + ".users",
	Path:         dashboardRoutePrefix + "/users",
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "UsersList",
	Middleware: []string{
		"AuthOnly",
	},
}

var DashboardUserEdit = Route{
	Name:         dashboardNamePrefix + ".user.edit",
	Path:         dashboardRoutePrefix + "/users/:id/edit",
	Method:       http.MethodGet,
	Handler:      "Dashboard",
	HandleMethod: "EditUser",
	Middleware: []string{
		"AuthOnly",
	},
}

var DashboardUserUpdate = Route{
	Name:         dashboardNamePrefix + ".user.update",
	Path:         dashboardRoutePrefix + "/users/:id",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "UpdateUser",
	Middleware: []string{
		"AuthOnly",
	},
}

var DashboardUserDelete = Route{
	Name:         dashboardNamePrefix + ".user.delete",
	Path:         dashboardRoutePrefix + "/users/:id/delete",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "DeleteUser",
	Middleware: []string{
		"AuthOnly",
	},
}

var DashboardUserToggleAdmin = Route{
	Name:         dashboardNamePrefix + ".user.toggle_admin",
	Path:         dashboardRoutePrefix + "/users/:id/toggle-admin",
	Method:       http.MethodPost,
	Handler:      "Dashboard",
	HandleMethod: "MakeUserAdmin",
	Middleware: []string{
		"AuthOnly",
	},
}
