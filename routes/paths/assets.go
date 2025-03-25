package paths

import "github.com/mbvlabs/grafto/static"

const assetPrefix = "/assets"

var (
	Robots = register(
		Path{
			Name: "assets.robots",
			URL:  "/robots.txt",
		},
	)

	Sitemap = register(
		Path{
			Name: "assets.sitemap",
			URL:  "/sitemap.xml",
		},
	)

	MainCss = register(
		Path{
			Name: "assets.main_css",
			URL: func() string {
				return assetPrefix + "/" + static.MainCssFile
			}(),
		},
	)

	HtmxJS = register(
		Path{
			Name: "assets.htmx",
			URL:  assetPrefix + "/" + "htmx-2_0_4.min.js",
		},
	)

	AlpineJS = register(
		Path{
			Name: "assets.alpine",
			URL:  assetPrefix + "/" + "alpine-3_14_8.min.js",
		},
	)

	Favicon = register(
		Path{
			Name: "assets.favicon",
			URL:  "/favicon.ico",
		},
	)
	Favicon16 = register(
		Path{
			Name: "assets.favicon_16_16",
			URL:  assetPrefix + "/" + "favicon-16x16.png",
		},
	)
	Favicon32 = register(
		Path{
			Name: "assets.favicon_32_32",
			URL:  assetPrefix + "/" + "favicon-32x32.png",
		},
	)
)
