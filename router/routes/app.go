package routes

import (
	"fmt"
	"net/http"
)

const appNamePrefix = "app"

var App = []Route{
	LandingPage,
	AboutPage,
	Redirect.Route,
}

var LandingPage = Route{
	Name:         appNamePrefix + ".landing_page",
	Path:         "/",
	Method:       http.MethodGet,
	Handler:      "Pages",
	HandleMethod: "LandingPage",
}

var AboutPage = Route{
	Name:         appNamePrefix + ".about_page",
	Path:         "/about",
	Method:       http.MethodGet,
	Handler:      "Pages",
	HandleMethod: "AboutPage",
}

var Redirect = redirect{
	Route: Route{
		Name:         appNamePrefix + ".redirect",
		Path:         "/redirect",
		Handler:      "Pages",
		HandleMethod: "Redirect",
		Method:       http.MethodGet,
	},
}

type redirect struct {
	Route
}

func (r redirect) WithQuery(route Route) string {
	return fmt.Sprintf("%s?to=%s", r.Path, route.Path)
}
