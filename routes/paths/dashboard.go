package paths

import "fmt"

const dashboardPrefix = "/dashboard"

var DashboardHome = register(
	Path{Name: "dashboard.home", URL: fmt.Sprintf("%s/", dashboardPrefix)},
)
