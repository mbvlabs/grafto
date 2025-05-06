package routes

import (
	"net/http"
)

const (
	assetsRoutePrefix = "/assets"
	assetsNamePrefix  = "assets"
)

var Assets = []Route{
	Robots,
	Sitemap,
	CssEntrypoint,
	AllCss,
	JsEntrypoint,
	AllJs,
	// AlpineJS,
	// Htmx,
	Favicon16,
	Favicon32,
}

var Robots = Route{
	Name:     assetsNamePrefix + ".robots",
	Path:     assetsNamePrefix + "/robots.txt",
	Method:   http.MethodGet,
	CtrlName: "Robots",
}

var Sitemap = Route{
	Name:     assetsNamePrefix + ".sitemap",
	Path:     assetsNamePrefix + "/sitemap.xml",
	Method:   http.MethodGet,
	CtrlName: "Sitemap",
}

var CssEntrypoint = Route{
	Name:     assetsNamePrefix + "css.entry",
	Path:     assetsNamePrefix + "/css/styles.css",
	Method:   http.MethodGet,
	CtrlName: "Styles",
}

var AllCss = Route{
	Name:     assetsNamePrefix + "css.all",
	Path:     assetsNamePrefix + "/css/:file",
	Method:   http.MethodGet,
	CtrlName: "AllCss",
}

var JsEntrypoint = Route{
	Name:     assetsNamePrefix + "js.entry",
	Path:     assetsNamePrefix + "/js/script.js",
	Method:   http.MethodGet,
	CtrlName: "Scripts",
}

var AllJs = Route{
	Name:     assetsNamePrefix + "js.all",
	Path:     assetsNamePrefix + "/js/:file",
	Method:   http.MethodGet,
	CtrlName: "AllJs",
}

// var Htmx = Route{
// 	Name:     assetsNamePrefix + ".htmx",
// 	Path:     assetsNamePrefix + "/htmx",
// 	Method:   http.MethodGet,
// 	CtrlName: "Htmx",
// }
//
// var AlpineJS = Route{
// 	Name:     assetsNamePrefix + ".alpine",
// 	Path:     assetsNamePrefix + "/alpine",
// 	Method:   http.MethodGet,
// 	CtrlName: "AlpineJS",
// }

var Favicon16 = Route{
	Name:     assetsNamePrefix + ".favicon_16",
	Path:     assetsNamePrefix + "/favicon",
	Method:   http.MethodGet,
	CtrlName: "Favicon16",
}

var Favicon32 = Route{
	Name:     assetsNamePrefix + ".favicon_32",
	Path:     assetsNamePrefix + "/favicon",
	Method:   http.MethodGet,
	CtrlName: "Favicon32",
}
