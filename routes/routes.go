package routes

type RouteGroup struct{}

type Router struct {
	Api RouteGroup
	Web RouteGroup
}
