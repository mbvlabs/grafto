package routes

type Route struct {
	Name        string
	Path        string
	HandlerName string
	Method      string
	Middleware  []string
}

var AllRoutes = func() []Route {
	var r []Route
	r = append(r, ApiV1...)
	r = append(r, App...)
	r = append(r, Authentication...)
	r = append(r, Dashboard...)
	r = append(r, Registration...)

	return r
}()
