package routes

type Route struct {
	Name       string
	Path       string
	CtrlName   string
	Method     string
	Middleware []string
}
