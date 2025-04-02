package paths

const dashboardPrefix = "/dashboard"

var DashboardHome = register(
	Path{Name: "dashboard.home", URL: dashboardPrefix},
)
