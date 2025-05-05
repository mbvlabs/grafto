package routes

import (
	"net/http"
)

const appNamePrefix = "auth"

var App = []Route{
	LandingPage,
	AboutPage,
}

var LandingPage = Route{
	Name:     appNamePrefix + ".landing_page",
	Path:     "/",
	Method:   http.MethodGet,
	CtrlName: "LandingPage",
}

var AboutPage = Route{
	Name:     appNamePrefix + ".about_page",
	Path:     "/about",
	Method:   http.MethodGet,
	CtrlName: "AboutPage",
}
