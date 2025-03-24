package paths

import "fmt"

const assetPrefix = "/assets"

var (
	Robots = register(
		Path{
			Name: "assets.robots",
			URL:  fmt.Sprintf("%s/robots.txt", assetPrefix),
		},
	)
	Sitemap = register(
		Path{
			Name: "assets.sitemap",
			URL:  fmt.Sprintf("%s/sitemap.xml", assetPrefix),
		},
	)
	Favicon = register(
		Path{
			Name: "assets.favicon",
			URL:  fmt.Sprintf("%s/favicon.ico", assetPrefix),
		},
	)
)
