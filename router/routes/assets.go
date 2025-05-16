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
	Favicon16,
	Favicon32,
}

var Robots = Route{
	Name:        assetsNamePrefix + ".robots",
	Path:        assetsNamePrefix + "/robots.txt",
	Method:      http.MethodGet,
	HandlerName: "Robots",
}

var Sitemap = Route{
	Name:        assetsNamePrefix + ".sitemap",
	Path:        assetsNamePrefix + "/sitemap.xml",
	Method:      http.MethodGet,
	HandlerName: "Sitemap",
}

var CssEntrypoint = Route{
	Name:        assetsNamePrefix + "css.entry",
	Path:        assetsNamePrefix + "/css/styles.css",
	Method:      http.MethodGet,
	HandlerName: "Styles",
}

var AllCss = Route{
	Name:        assetsNamePrefix + "css.all",
	Path:        assetsNamePrefix + "/css/:file",
	Method:      http.MethodGet,
	HandlerName: "AllCss",
}

var JsEntrypoint = Route{
	Name:        assetsNamePrefix + "js.entry",
	Path:        assetsNamePrefix + "/js/script.js",
	Method:      http.MethodGet,
	HandlerName: "Scripts",
}

var AllJs = Route{
	Name:        assetsNamePrefix + "js.all",
	Path:        assetsNamePrefix + "/js/:file",
	Method:      http.MethodGet,
	HandlerName: "AllJs",
}

var Favicon16 = Route{
	Name:        assetsNamePrefix + ".favicon_16",
	Path:        assetsNamePrefix + "/favicon",
	Method:      http.MethodGet,
	HandlerName: "Favicon16",
}

var Favicon32 = Route{
	Name:        assetsNamePrefix + ".favicon_32",
	Path:        assetsNamePrefix + "/favicon",
	Method:      http.MethodGet,
	HandlerName: "Favicon32",
}
