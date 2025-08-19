package routes

import (
	"net/http"
)

const (
	assetsRoutePrefix = "/assets"
	assetsNamePrefix  = "assets"
)

var Robots = Route{
	Name:         assetsNamePrefix + ".robots",
	Path:         assetsNamePrefix + "/robots.txt",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Robots",
}

var Sitemap = Route{
	Name:         assetsNamePrefix + ".sitemap",
	Path:         assetsNamePrefix + "/sitemap.xml",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Sitemap",
}

var CssEntrypoint = Route{
	Name:         assetsNamePrefix + "css.entry",
	Path:         assetsNamePrefix + "/css/:version/styles.css",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Styles",
}

var AllCss = Route{
	Name:         assetsNamePrefix + "css.all",
	Path:         assetsNamePrefix + "/css/:version/:file",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "AllCss",
}

var JsEntrypoint = Route{
	Name:         assetsNamePrefix + "js.entry",
	Path:         assetsNamePrefix + "/js/:version/script.js",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Scripts",
}

var AllJs = Route{
	Name:         assetsNamePrefix + "js.all",
	Path:         assetsNamePrefix + "/js/:version/:file",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "AllJs",
}

var Favicon16 = Route{
	Name:         assetsNamePrefix + ".favicon_16",
	Path:         assetsNamePrefix + "/favicon",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Favicon16",
}

var Favicon32 = Route{
	Name:         assetsNamePrefix + ".favicon_32",
	Path:         assetsNamePrefix + "/favicon",
	Method:       http.MethodGet,
	Handler:      "Assets",
	HandleMethod: "Favicon32",
}
