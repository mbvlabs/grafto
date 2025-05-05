package routes

import "net/http"

const (
	assetsRoutePrefix = "/assets"
	assetsNamePrefix  = "assets"
)

var Assets = []Route{
	Robots,
	Sitemap,
	Bootstrap,
	MainCss,
	AlpineJS,
	Htmx,
}

var Robots = Route{
	Name:     assetsNamePrefix + ".robots",
	Path:     "/robots.txt",
	Method:   http.MethodGet,
	CtrlName: "Robots",
}

var Sitemap = Route{
	Name:     assetsNamePrefix + ".sitemap",
	Path:     "/sitemap.xml",
	Method:   http.MethodGet,
	CtrlName: "Sitemap",
}

var Bootstrap = Route{
	Name:     assetsNamePrefix + ".bootstrap_grid",
	Path:     "/bootstrap-v5_3_3.min.css",
	Method:   http.MethodGet,
	CtrlName: "BootstrapGrid",
}

var MainCss = Route{
	Name:     assetsNamePrefix + ".styles",
	Path:     "/styles.css",
	Method:   http.MethodGet,
	CtrlName: "MainCss",
}

var Htmx = Route{
	Name:     assetsNamePrefix + ".htmx",
	Path:     "/htmx",
	Method:   http.MethodGet,
	CtrlName: "Htmx",
}

var AlpineJS = Route{
	Name:     assetsNamePrefix + ".alpine",
	Path:     "/alpine",
	Method:   http.MethodGet,
	CtrlName: "AlpineJS",
}
