package paths

var (
	Redirect = register(
		Path{
			Name: "app.redirect",
			URL:  "/redirect",
		},
	)
	LandingPage = register(
		Path{
			Name: "app.landing_page",
			URL:  "/",
		},
	)
	About = register(
		Path{
			Name: "app.home",
			URL:  "/about",
		},
	)
)
