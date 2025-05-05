package routes

import (
	"net/http"
)

var App = []Route{
	LandingPage,
	AboutPage,
}

var LandingPage = Route{
	Name:     "app.landing_page",
	Path:     "/",
	Method:   http.MethodGet,
	CtrlName: "LandingPage",
}

var AboutPage = Route{
	Name:     "app.about_page",
	Path:     "/about",
	Method:   http.MethodGet,
	CtrlName: "AboutPage",
}
