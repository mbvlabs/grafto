package paths

import (
	"github.com/mbvlabs/grafto/static"
)

const AssetPrefix = "/assets"

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

	BootstrapGrid = register(
		Path{
			Name: "assets.bootstrap_grid",
			URL: func() string {
				return AssetPrefix + "/" + "bootstrap-v5_3_3.min.css"
			}(),
		},
	)

	NewCss = register(
		Path{
			Name: "assets.new_css",
			URL: func() string {
				return AssetPrefix + "/" + "styles.css"
			}(),
		},
	)

	NormalizeCss = register(
		Path{
			Name: "assets.normalize_css",
			URL: func() string {
				return AssetPrefix + "/" + "normalize.css"
			}(),
		},
	)

	LayoutCss = register(
		Path{
			Name: "assets.layout_css",
			URL: func() string {
				return AssetPrefix + "/" + "layout.css"
			}(),
		},
	)

	BaseCss = register(
		Path{
			Name: "assets.base_css",
			URL: func() string {
				return AssetPrefix + "/" + "base.css"
			}(),
		},
	)

	NavCss = register(
		Path{
			Name: "assets.nav_css",
			URL: func() string {
				return AssetPrefix + "/" + "nav.css"
			}(),
		},
	)

	MainCss = register(
		Path{
			Name: "assets.main_css",
			URL: func() string {
				return AssetPrefix + "/" + static.MainCssFile
			}(),
		},
	)

	HtmxJS = register(
		Path{
			Name: "assets.htmx",
			URL:  AssetPrefix + "/" + "htmx-2_0_4.min.js",
		},
	)

	AlpineJS = register(
		Path{
			Name: "assets.alpine",
			URL:  AssetPrefix + "/" + "alpine-3_14_8.min.js",
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
			URL:  AssetPrefix + "/" + "favicon-16x16.png",
		},
	)
	Favicon32 = register(
		Path{
			Name: "assets.favicon_32_32",
			URL:  AssetPrefix + "/" + "favicon-32x32.png",
		},
	)
)
