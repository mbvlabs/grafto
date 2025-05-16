package routes

type Route struct {
	Name        string
	Path        string
	HandlerName string
	Method      string
	Middleware  []string
}
